# AlphaSignal Telegram Bot

AlphaSignal API'den haberleri çeken, `_id` ve `timestamp` değişimini takip eden ve yeni içerik geldiğinde Telegram kanalına veya abone gruplara gönderen Go tabanlı bot.

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
| `STATE_FILE` | State dosya yolu (varsayılan: `./state.json`) |
| `ALPHASIGNAL_API` | API URL (varsayılan: `https://alphasignal.ai/api/last-campaign`) |

## Kullanım

**Sürekli çalışma (30 dakikada bir kontrol):**
```bash
./ai-news-bot
```

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

İlk çalıştırmada state kaydedilir ancak mesaj gönderilmez (spam önleme). Sonraki çalışmalarda `_id` veya `timestamp` değiştiğinde yeni haberler gönderilir.

## Cron Örneği

```cron
*/30 * * * * cd /path/to/ai-news-bot && ./ai-news-bot -once
```
