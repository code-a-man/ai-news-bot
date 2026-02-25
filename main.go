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

	run := func() {
		if err := runCheck(cfg); err != nil {
			log.Printf("check error: %v", err)
		}
	}

	if *once || *dryRun {
		if *dryRun {
			runDryRun(cfg)
		} else {
			run()
		}
		return
	}

	c := cron.New()
	if _, err := c.AddFunc("*/30 * * * *", run); err != nil {
		log.Fatalf("cron: %v", err)
	}
	c.Start()
	log.Println("cron started: every 30 minutes")

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
	log.Println("--- formatted message ---")
	log.Println(text)
	log.Println("--- end ---")
}
