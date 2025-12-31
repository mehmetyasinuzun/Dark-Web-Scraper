# TOR Scraper - Proje Raporu

**Proje:** Dark Web Scraper  
**Tarih:** 31 Aralık 2025  
**GitHub:** https://github.com/mehmetyasinuzun/Dark-Web-Scraper

---

## 1. Giriş

Bu raporda Tor ağı üzerinden çalışan bir web scraper aracının nasıl geliştirildiğini anlatıyorum. Projenin amacı, .onion uzantılı siteleri otomatik olarak tarayıp içeriklerini kaydetmek. CTI (Cyber Threat Intelligence) alanında bu tür araçlar, dark web'deki aktiviteleri izlemek için kullanılıyor.

Normalde bu sitelere tek tek girip bakmak gerekiyor ama yüzlerce site olunca bu iş imkansız hale geliyor. Bu yüzden otomatik bir araç yazmak mantıklıydı.

---

## 2. Kullandığım Teknolojiler

### 2.1 Programlama Dili: Go

Projeyi Go ile yazdım. Go'yu seçmemin birkaç sebebi var:

- Derlenmiş bir dil, tek bir exe dosyası çıkıyor
- HTTP işlemleri için güzel kütüphaneleri var
- Öğrenmesi nispeten kolay

### 2.2 Kütüphaneler

| Kütüphane | Ne için kullandım |
|-----------|-------------------|
| net/http | HTTP istekleri atmak için |
| golang.org/x/net/proxy | Tor'a bağlanmak için (SOCKS5 proxy) |
| gopkg.in/yaml.v3 | targets.yaml dosyasını okumak için |
| github.com/chromedp/chromedp | Screenshot almak için |
| os, io, fmt, time, strings, context | Genel işlemler |

### 2.3 Tor Browser

Tor Browser arka planda çalışırken 9150 portunda bir SOCKS5 proxy açıyor. Programım bu proxy üzerinden internete çıkıyor, böylece tüm trafik Tor ağından geçiyor ve .onion sitelerine erişebiliyorum.

---

## 3. Proje Yapısı

```
Dark-Web-Scraper/
├── main.go           # Ana kod
├── targets.yaml      # Taranacak sitelerin listesi
├── go.mod            # Go modül tanımı
├── go.sum            # Bağımlılık kontrol dosyası
├── README.md         # Kullanım kılavuzu
├── RAPOR.md          # Bu rapor
├── .gitignore        # Git'in yoksayacağı dosyalar
└── output/           # Çıktıların kaydedildiği klasör
    ├── html/         # HTML dosyaları
    ├── screenshots/  # Ekran görüntüleri
    └── scan_report_*.log  # Tarama raporları
```

---

## 4. Kodun Detaylı Açıklaması

### 4.1 Veri Yapıları

İlk önce YAML dosyasını okuyabilmek için struct'lar tanımladım:

```go
type Target struct {
    Name string `yaml:"name"`
    Url  string `yaml:"url"`
}

type Config struct {
    Forums []Target `yaml:"forums"`
}
```

Target struct'ı her bir siteyi temsil ediyor. Name site adı, Url ise .onion adresi. Config struct'ı ise tüm hedefleri bir arada tutuyor.

`yaml:"name"` kısmı Go'ya "YAML'daki name alanını bu değişkene ata" diyor. Buna struct tag deniyor.

### 4.2 YAML Dosyasını Okuma

```go
data, err := os.ReadFile("targets.yaml")
if err != nil {
    fmt.Println("YAML dosyası okunamadı: ", err)
    return
}

config := Config{}
err = yaml.Unmarshal(data, &config)
```

Önce dosyayı okuyorum, sonra yaml.Unmarshal ile parse ediyorum. &config kısmındaki & işareti config değişkeninin bellek adresini veriyor, böylece Unmarshal fonksiyonu doğrudan o değişkene yazabiliyor.

### 4.3 Tor'a Bağlanma

```go
func createTorClient() (*http.Client, error) {
    torProxy := "127.0.0.1:9150"
    
    dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
    if err != nil {
        fmt.Println("Tor bağlantısı kurulamadı: ", err)
        return nil, err
    }
    
    transport := &http.Transport{
        Dial: dialer.Dial,
    }
    
    client := &http.Client{
        Transport: transport,
        Timeout:   60 * time.Second,
    }
    return client, nil
}
```

Bu fonksiyon Tor proxy'sini kullanan özel bir HTTP client oluşturuyor. Adım adım:

1. `proxy.SOCKS5()` ile Tor'un proxy'sine bağlanan bir dialer oluşturuyorum
2. Bu dialer'ı kullanan bir Transport oluşturuyorum
3. Transport'u kullanan bir HTTP Client oluşturuyorum
4. Timeout'u 60 saniye yaptım çünkü Tor yavaş olabiliyor

### 4.4 Site Tarama

```go
func scrapeSite(client *http.Client, url string) (string, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", err
    }
    
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
```

Bu fonksiyon verilen URL'e istek atıp HTML içeriğini döndürüyor.

User-Agent header'ı ekledim çünkü bazı siteler bot gibi görünen istekleri engelliyor. Chrome gibi görünmek için gerçek bir Chrome User-Agent'ı kullandım.

`defer resp.Body.Close()` satırı önemli. defer kelimesi "fonksiyon bitince bunu çalıştır" demek. Body'yi kapatmazsak bellek sızıntısı olur.

### 4.5 Screenshot Alma

```go
func screenshotAl(url string, kayitYolu string) error {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.ProxyServer("socks5://127.0.0.1:9150"),
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
    )
    
    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    defer cancel()
    
    ctx, cancel := chromedp.NewContext(allocCtx)
    defer cancel()
    
    ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
    defer cancel()
    
    var buf []byte
    err := chromedp.Run(ctx,
        chromedp.Navigate(url),
        chromedp.Sleep(5*time.Second),
        chromedp.FullScreenshot(&buf, 90),
    )
    
    if err != nil {
        return err
    }
    
    return os.WriteFile(kayitYolu, buf, 0644)
}
```

chromedp kütüphanesi Chrome'u programatik olarak kontrol etmemi sağlıyor. Headless modda çalışıyor yani ekranda pencere açılmıyor.

Proxy ayarını burada da yaptım ki screenshot alırken de Tor üzerinden gitsin.

90 saniye timeout koydum çünkü Tor üzerinden sayfa yüklemek uzun sürebiliyor. 5 saniye Sleep koydum ki sayfa tam yüklensin.

### 4.6 Ana Döngü

```go
for i, target := range config.Forums {
    fmt.Printf("\n[%d/%d] %s taranıyor...\n", i+1, len(config.Forums), target.Name)
    
    html, err := scrapeSite(client, target.Url)
    
    if err != nil {
        fmt.Printf("[HATA] %s: %v\n", target.Name, err)
        rapor += "[HATA] " + target.Name + " - " + err.Error() + "\n"
        basarisiz++
    } else {
        fmt.Printf("[OK] %s - %d byte veri alındı\n", target.Name, len(html))
        
        // HTML kaydet
        // Screenshot al
        // Rapora ekle
        basarili++
    }
}
```

Tüm siteleri sırayla tarıyorum. Hata olursa programı durdurmuyorum, sadece hatayı not edip sonraki siteye geçiyorum. Bu önemli çünkü dark web'de siteler sürekli kapanıp açılıyor, bir site çalışmıyor diye tüm tarama durmamalı.

### 4.7 Dosya İsimlendirme

```go
tarihSaat := time.Now().Format("02-01-2006_15-04-05")
htmlYol := "output/html/" + dosyaAdi + "_" + tarihSaat + ".html"
```

Her dosyanın sonuna tarih ve saat ekliyorum. Böylece programı birden fazla kez çalıştırsam dosyalar üst üste binmiyor, hepsini ayrı ayrı görebiliyorum.

---

## 5. targets.yaml İçeriği

```yaml
forums:
  - name: "DuckDuckGo Onion"
    url: "https://duckduckgogg42xjoc72x3sjasowoarfbgcmvfimaftt6twagswzczad.onion"
  - name: "Ahmia Search"
    url: "http://juhanurmihxlp77nkq76byazcldy2hlmovfu2epvl5ankdibsot4csyd.onion"
  - name: "Tor Project"
    url: "http://2gzyxa5ihm7nsggfxnu52rck2vv4rvmdlkiu3ez2tuj35pzdll5ncqad.onion"
  - name: "CIA"
    url: "http://ciadotgov4sjwlzihbbgxnqg3xiyrg7so2r2o3lt5wz5ypk4sxyjstad.onion"
  - name: "Stuff Forum"
    url: "http://stuffdo2micuw3krnffqxb7ldc66pkfsindy5stobovme5ywo4susoad.onion/forum/"
  - name: "PGP Shop Forum"
    url: "http://pgpshopyoohxel4jen5trjfnenou7sodhabd37v2a46hmjhfdxwntjad.onion/pgp/?product_tag=forum"
  - name: "SuprBay: The PirateBay Forum"
    url: "http://suprbaydvdcaynfo4dgdzgxb4zuso7rftlil5yg5kqjefnw4wq4ulcad.onion/"
  - name: "Darknet Forum"
    url: "http://darknet3jmlenxbn5tbqrd6tbjbuqipxp7lekxbl7fycpxzaohnsduid.onion/"
  - name: "InfoCon Hacking and Security Conference Archives"
    url: "http://w27irt6ldaydjoacyovepuzlethuoypazhhbot6tljuywy52emetn7qd.onion/"
```

Toplamda 9 site var. Bunların bir kısmı arama motorları (DuckDuckGo, Ahmia), bir kısmı resmi siteler (Tor Project, CIA), bir kısmı da forumlar.

---

## 6. Çıktı Örnekleri

### 6.1 Konsol Çıktısı

```
=== Thor'un Scraper'ı ===
Tarih:  31-12-2025_22-53-58
Hedefler Yüklendi:  9  adet
1 .  DuckDuckGo Onion
 -  https://duckduckgogg42xjoc72x3sjasowoarfbgcmvfimaftt6twagswzczad.onion
2 .  Ahmia Search
 -  http://juhanurmihxlp77nkq76byazcldy2hlmovfu2epvl5ankdibsot4csyd.onion
...
Tor Bağlantısı Başarılı
HTTP Durum Kodu:  200

=== Tarama Başlıyor ===

[1/9] DuckDuckGo Onion taranıyor...
[OK] DuckDuckGo Onion - 72442 byte veri alındı
HTML kaydedildi:  output/html/DuckDuckGo_Onion_31-12-2025_22-53-58.html
Screenshot kaydedildi:  output/screenshots/DuckDuckGo_Onion_31-12-2025_22-53-58.png

[2/9] Ahmia Search taranıyor...
[OK] Ahmia Search - 15234 byte veri alındı
...
```

### 6.2 scan_report.log Örneği

```
=== TARAMA RAPORU ===
Tarih: 31-12-2025_22-53-58

[OK] DuckDuckGo Onion
  URL: https://duckduckgogg42xjoc72x3sjasowoarfbgcmvfimaftt6twagswzczad.onion
  HTML: output/html/DuckDuckGo_Onion_31-12-2025_22-53-58.html
  Screenshot: output/screenshots/DuckDuckGo_Onion_31-12-2025_22-53-58.png

[OK] Ahmia Search
  URL: http://juhanurmihxlp77nkq76byazcldy2hlmovfu2epvl5ankdibsot4csyd.onion
  HTML: output/html/Ahmia_Search_31-12-2025_22-53-58.html
  Screenshot: output/screenshots/Ahmia_Search_31-12-2025_22-53-58.png

...

=== OZET ===
Basarili: 8
Basarisiz: 1
Toplam: 9
```

---

## 7. Karşılaştığım Sorunlar ve Çözümler

### 7.1 Tor Yavaşlığı

Tor ağı normal internetten çok daha yavaş. İlk başta 30 saniye timeout koymuştum ama bazı siteler yüklenemiyordu. 60 saniyeye çıkardım, screenshot için 90 saniye yaptım.

### 7.2 Dead Linkler

Dark web'de siteler sürekli kapanıyor. Bir site kapanınca program hata veriyor ama ben bunu yakalayıp devam ettiriyorum. Böylece bir site çalışmasa bile diğerleri taranabiliyor.

### 7.3 User-Agent

Bazı siteler Go'nun varsayılan User-Agent'ını görünce istek atmayı engelliyor. Chrome User-Agent'ı ekleyince bu sorun çözüldü.

### 7.4 Screenshot Zamanlaması

chromedp sayfayı açar açmaz screenshot alıyordu ama sayfa henüz yüklenmemiş oluyordu. 5 saniye Sleep ekleyince düzeldi.

---

## 8. Kurulum ve Çalıştırma

### 8.1 Gereksinimler

- Go 1.20 veya üstü
- Tor Browser
- Chrome veya Chromium (screenshot için)

### 8.2 Kurulum

```bash
git clone https://github.com/mehmetyasinuzun/Dark-Web-Scraper.git
cd Dark-Web-Scraper
go mod tidy
```

### 8.3 Çalıştırma

1. Tor Browser'ı aç ve arka planda çalışır bırak
2. Terminalde `go run main.go` yaz
3. Çıktılar output/ klasörüne düşecek

---

## 9. Sonuç

Proje istenen tüm özellikleri karşılıyor:

| Özellik | Durum |
|---------|-------|
| Go dili kullanımı | ✓ |
| YAML'dan hedef okuma | ✓ |
| Tor proxy desteği | ✓ |
| HTTP istekleri | ✓ |
| User-Agent | ✓ |
| Hata yönetimi | ✓ |
| HTML kaydetme | ✓ |
| Screenshot alma | ✓ |
| Tarih damgalı dosyalar | ✓ |
| Tarama raporu | ✓ |

Program basit tutmaya çalıştım. Gereksiz karmaşıklık eklemedim, ne lazımsa onu yazdım. Kod okunabilir ve anlaşılır durumda.

---

## 10. Kaynaklar

- Go dökümantasyonu: https://go.dev/doc/
- chromedp: https://github.com/chromedp/chromedp
- Tor Project: https://www.torproject.org/
