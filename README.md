# AI News Telegram Bot

AlphaSignal haberleri ve Claude Status RSS'i takip eden Telegram botu.

- **AlphaSignal**: `_id`/`timestamp` değişince yeni haberleri gönderir
- **Claude Status**: Yeni incident gelince mesaj atar, güncelleme olunca aynı mesajı düzenler

## Kurulum

```bash
go build -o ai-news-bot .
```

## Konfigürasyon

Ortam değişkenleri:

| Değişken | Açıklama |
|----------|----------|
| `TELEGRAM_BOT_TOKEN` | BotFather'dan alınan token (zorunlu) |
| `TELEGRAM_CHAT_IDS` | Hedef chat ID'leri, virgülle ayrılmış (örn: `@channel,-1001234567890`) |
| `STATE_FILE` | AlphaSignal state (varsayılan: `./state.json`) |
| `ALPHASIGNAL_API` | AlphaSignal API URL |
| `CLAUDE_STATUS_RSS_URL` | Claude Status RSS (varsayılan: `https://status.claude.com/history.rss`) |
| `STATE_RSS_FILE` | Claude Status state (varsayılan: `./state_rss.json`) |

## Kullanım

**Sürekli çalışma (iç cron):**
```bash
./ai-news-bot
```

- **AlphaSignal (AI News):** saatte 1 kez
- **Claude Status:** 10 dakikada 1 kez

**Tek seferlik (cron için):**
```bash
./ai-news-bot -once
```

**Test (fetch + parse, gönderme yok):**
```bash
./ai-news-bot -dry-run
```

## Telegram Chat ID

- **Kanal**: `@channel_username` veya `-100xxxxxxxxxx`
- **Grup**: Botu gruba ekleyin, chat ID negatif sayı (örn: `-1001234567890`)

## İlk Çalıştırma

- **AlphaSignal**: İlk çalıştırmada state kaydedilir, mesaj gönderilmez. Sonraki çalışmalarda `_id`/`timestamp` değişince haberler gönderilir.
- **Claude Status**: İlk çalıştırmada yalnızca açık (`Resolved` olmayan) incident'ler için mesaj gönderilir ve message ID'leri state'e kaydedilir. Kapalı incident'ler state'e işlenir ama mesaj atılmaz. Sonraki çalışmalarda yeni açık incident mesajı atılır, aynı incident güncellendikçe mevcut mesaj düzenlenir.

## Cron Örneği

```cron
*/30 * * * * cd /path/to/ai-news-bot && ./ai-news-bot -once
```
