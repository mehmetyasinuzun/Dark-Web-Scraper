package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v3"
)

type Target struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

type Config struct {
	Forums []Target `yaml:"forums"`
}

func main() {
	fmt.Println("=== Thor'un Scraper'ı ===")

	tarihSaat := time.Now().Format("02-01-2006_15-04-05")
	fmt.Println("Tarih: ", tarihSaat)

	os.MkdirAll("output/html", 0755)
	os.MkdirAll("output/screenshots", 0755)

	data, err := os.ReadFile("targets.yaml")
	if err != nil {
		fmt.Println("YAML dosyası okunamadı: ", err)
		return
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Parse Hatası: ", err)
		return
	}

	fmt.Println("Hedefler Yüklendi: ", len(config.Forums), " adet")

	for i, t := range config.Forums {

		fmt.Println(i+1, ". ", t.Name, "\n - ", t.Url)

	}

	client, err := createTorClient()
	if err != nil {
		fmt.Println("Tor bağlantısı oluşturulamadı")
		return
	}

	resp, err := client.Get("https://check.torproject.org")
	if err != nil {
		fmt.Println("Tor uzerinden istek gonderilemedi: ", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Tor Bağlantısı Başarılı")
	fmt.Println("HTTP Durum Kodu: ", resp.StatusCode)

	fmt.Println("\n=== Tarama Başlıyor ===")

	rapor := "=== TARAMA RAPORU ===\n"
	rapor += "Tarih: " + tarihSaat + "\n\n"

	basarili := 0
	basarisiz := 0

	for i, target := range config.Forums {
		fmt.Printf("\n[%d/%d] %s taranıyor...\n", i+1, len(config.Forums), target.Name)

		html, err := scrapeSite(client, target.Url)

		if err != nil {
			fmt.Printf("[HATA] %s: %v\n", target.Name, err)
			rapor += "[HATA] " + target.Name + " - " + err.Error() + "\n"
			basarisiz++
		} else {
			fmt.Printf("[OK] %s - %d byte veri alındı\n", target.Name, len(html))

			dosyaAdi := strings.ReplaceAll(target.Name, " ", "_")
			dosyaAdi = strings.ReplaceAll(dosyaAdi, ":", "")

			htmlYol := "output/html/" + dosyaAdi + "_" + tarihSaat + ".html"
			os.WriteFile(htmlYol, []byte(html), 0644)
			fmt.Println("HTML kaydedildi: ", htmlYol)

			screenshotYol := "output/screenshots/" + dosyaAdi + "_" + tarihSaat + ".png"
			ssErr := screenshotAl(target.Url, screenshotYol)
			if ssErr != nil {
				fmt.Println("Screenshot alinamadi: ", ssErr)
			} else {
				fmt.Println("Screenshot kaydedildi: ", screenshotYol)
			}

			rapor += "[OK] " + target.Name + "\n"
			rapor += "  URL: " + target.Url + "\n"
			rapor += "  HTML: " + htmlYol + "\n"
			rapor += "  Screenshot: " + screenshotYol + "\n\n"
			basarili++
		}
	}

	rapor += "\n=== OZET ===\n"
	rapor += fmt.Sprintf("Basarili: %d\n", basarili)
	rapor += fmt.Sprintf("Basarisiz: %d\n", basarisiz)
	rapor += fmt.Sprintf("Toplam: %d\n", basarili+basarisiz)

	os.WriteFile("output/scan_report_"+tarihSaat+".log", []byte(rapor), 0644)
	fmt.Println("\nRapor kaydedildi: output/scan_report_" + tarihSaat + ".log")

	fmt.Println("\n=== Tarama Tamamlandı ===")

}

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

func scrapeSite(client *http.Client, url string) (string, error) {

	resp, err := client.Get(url)
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
