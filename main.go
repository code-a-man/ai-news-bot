package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ai-news-bot/config"
	"ai-news-bot/fetcher"
	"ai-news-bot/parser"
	"ai-news-bot/rss"
	"ai-news-bot/state"
	"ai-news-bot/telegram"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	_ = godotenv.Load()

	once := flag.Bool("once", false, "Run once and exit (for cron)")
	dryRun := flag.Bool("dry-run", false, "Fetch and parse only, no send (for testing)")
	flag.Parse()

	cfg := config.Load()
	if !*dryRun {
		if cfg.TelegramToken == "" {
			log.Fatal("TELEGRAM_BOT_TOKEN is required")
		}
		if len(cfg.ChatIDs) == 0 {
			log.Fatal("TELEGRAM_CHAT_IDS is required (comma-separated)")
		}
	}

	runAlphaSignal := func() {
		if err := runCheck(cfg); err != nil {
			log.Printf("alphasignal error: %v", err)
		}
	}
	runClaudeStatus := func() {
		if err := runClaudeStatusCheck(cfg); err != nil {
			log.Printf("claude status error: %v", err)
		}
	}

	if *once || *dryRun {
		if *dryRun {
			runDryRun(cfg)
		} else {
			runAlphaSignal()
			runClaudeStatus()
		}
		return
	}

	c := cron.New()
	if _, err := c.AddFunc("0 * * * *", runAlphaSignal); err != nil {
		log.Fatalf("cron: %v", err)
	}
	if _, err := c.AddFunc("*/10 * * * *", runClaudeStatus); err != nil {
		log.Fatalf("cron: %v", err)
	}

	c.Start()
	log.Println("cron started: alphasignal every 1 hour, claude status every 10 minutes")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	c.Stop()
	log.Println("stopped")
}

func runCheck(cfg *config.Config) error {
	campaign, err := fetcher.Fetch(cfg.APIURL)
	if err != nil {
		return err
	}

	s, err := state.Load(cfg.StateFile)
	if err != nil {
		return err
	}

	firstRun := s.LastID == "" && s.LastTimestamp == ""
	if firstRun {
		if err := state.Save(cfg.StateFile, campaign.ID, campaign.Timestamp); err != nil {
			return err
		}
		log.Println("first run: state saved, no message sent")
		return nil
	}

	if !s.HasChanged(campaign.ID, campaign.Timestamp) {
		log.Println("no new content")
		return nil
	}

	items, readTime, err := parser.ParseSummary(campaign.HTML)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		log.Println("no summary items parsed")
		_ = state.Save(cfg.StateFile, campaign.ID, campaign.Timestamp)
		return nil
	}

	text := parser.FormatMessage(items, readTime, campaign.Subject)
	tg, err := telegram.New(cfg.TelegramToken)
	if err != nil {
		return err
	}

	if err := tg.SendToChats(cfg.ChatIDs, text); err != nil {
		return err
	}

	if err := state.Save(cfg.StateFile, campaign.ID, campaign.Timestamp); err != nil {
		return err
	}
	log.Printf("sent to %d chat(s), campaign %s", len(cfg.ChatIDs), campaign.ID)
	return nil
}

func runClaudeStatusCheck(cfg *config.Config) error {
	items, err := rss.Fetch(cfg.ClaudeRSSURL)
	if err != nil {
		return err
	}

	s, err := state.LoadRSS(cfg.StateRSSFile)
	if err != nil {
		return err
	}

	tg, err := telegram.New(cfg.TelegramToken)
	if err != nil {
		return err
	}

	firstRun := len(s.Incidents) == 0
	if firstRun {
		for _, item := range items {
			guid := item.GUID
			if guid == "" {
				guid = item.Link
			}
			if guid == "" {
				continue
			}
			descHash := state.HashDescription(item.Description)
			isOpen := rss.IsOpenIncident(item)
			text := rss.FormatMessage(item)

			for _, chatID := range cfg.ChatIDs {
				if !isOpen {
					s.SetMessage(guid, chatID, 0, descHash)
					continue
				}

				sentID, err := tg.SendSingle(chatID, text)
				if err != nil {
					log.Printf("claude status first-run send failed: %v", err)
					continue
				}
				s.SetMessage(guid, chatID, sentID, descHash)
			}
		}
		if err := state.SaveRSS(cfg.StateRSSFile, s); err != nil {
			return err
		}
		log.Println("claude status: first run completed (open incidents sent, state saved)")
		return nil
	}

	modified := false
	for _, item := range items {
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}
		if guid == "" {
			continue
		}

		descHash := state.HashDescription(item.Description)
		text := rss.FormatMessage(item)
		isOpen := rss.IsOpenIncident(item)

		for _, chatID := range cfg.ChatIDs {
			msgID, lastHash, exists := s.GetMessage(guid, chatID)
			if !exists {
				if isOpen {
					sentID, err := tg.SendSingle(chatID, text)
					if err != nil {
						log.Printf("claude status send failed: %v", err)
						continue
					}
					s.SetMessage(guid, chatID, sentID, descHash)
				} else {
					s.SetMessage(guid, chatID, 0, descHash)
				}
				modified = true
			} else if lastHash != descHash {
				if msgID == 0 {
					if isOpen {
						sentID, err := tg.SendSingle(chatID, text)
						if err != nil {
							log.Printf("claude status send failed: %v", err)
							continue
						}
						s.SetMessage(guid, chatID, sentID, descHash)
					} else {
						s.SetMessage(guid, chatID, 0, descHash)
					}
				} else {
					if err := tg.EditMessage(chatID, msgID, text); err != nil {
						log.Printf("claude status edit failed: %v", err)
						continue
					}
					s.SetMessage(guid, chatID, msgID, descHash)
				}
				modified = true
			}
		}
	}

	if modified {
		if err := state.SaveRSS(cfg.StateRSSFile, s); err != nil {
			return err
		}
		log.Println("claude status: updated")
	}
	return nil
}

func runDryRun(cfg *config.Config) {
	campaign, err := fetcher.Fetch(cfg.APIURL)
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	log.Printf("fetched campaign %s (%s)", campaign.ID, campaign.Timestamp)

	items, readTime, err := parser.ParseSummary(campaign.HTML)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	log.Printf("parsed %d items, readTime=%q", len(items), readTime)

	text := parser.FormatMessage(items, readTime, campaign.Subject)
	log.Println("--- AlphaSignal ---")
	log.Println(text)
	log.Println("--- end ---")

	rssItems, err := rss.Fetch(cfg.ClaudeRSSURL)
	if err != nil {
		log.Printf("rss fetch: %v", err)
		return
	}
	log.Printf("--- Claude Status RSS (%d items) ---", len(rssItems))
	for i, item := range rssItems {
		if i >= 3 {
			log.Printf("... and %d more", len(rssItems)-3)
			break
		}
		log.Println(rss.FormatMessage(item))
		log.Println("---")
	}
}
