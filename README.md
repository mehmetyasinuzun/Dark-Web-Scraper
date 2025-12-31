# Dark Web Scraper

Tor ağı üzerinden .onion sitelerini tarayan bir araç. YAML dosyasından hedef listesi okuyup, her site için HTML kaydediyor ve screenshot alıyor.

## Ne işe yarıyor?

- `targets.yaml` dosyasındaki onion adreslerini sırayla tarıyor
- Tor proxy üzerinden bağlanıyor (SOCKS5)
- Her sitenin HTML içeriğini kaydediyor
- Screenshot alıyor
- Sonuçları `scan_report.log` dosyasına yazıyor

## Kurulum

```bash
git clone https://github.com/mehmetyasinuzun/Dark-Web-Scraper.git
cd Dark-Web-Scraper
go mod tidy
```

## Kullanım

Önce Tor Browser'ı aç ve arka planda çalışır durumda bırak. Sonra:

```bash
go run main.go
```

Çıktılar `output/` klasörüne düşüyor:
- `output/html/` - HTML dosyaları
- `output/screenshots/` - Ekran görüntüleri  
- `output/scan_report_TARIH.log` - Tarama raporu

## targets.yaml

Taranacak siteleri buraya ekliyorsun:

```yaml
forums:
  - name: "Site Adı"
    url: "http://xxxxx.onion"
```

## Gereksinimler

- Go 1.20+
- Tor Browser (9150 portunda çalışır durumda)
- Chrome/Chromium (screenshot için)

## Notlar

- Bazı siteler kapanmış veya yavaş olabilir, program takılmıyor devam ediyor
- Screenshot timeout 90 saniye, tor ağı yavaş olduğu için uzun tuttum
- Dead linkler raporda [HATA] olarak görünüyor
