package main

import (
	"encoding/json"
	"fmt"
	grab "github.com/cavaliercoder/grab/v3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

type GarStats struct {
	Version       int    `json:"VersionId"`
	GarXMLFullURL string `json:"GarXMLFullURL"`
	Date          string `json:"Date"`
}
type Conf struct {
	Db struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	}
}

func (c *Conf) GetConf() *Conf {

	yamlFile, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
func main() {
	var conf Conf
	conf.GetConf()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", conf.Db.Host, conf.Db.Port, conf.Db.User, conf.Db.Password, conf.Db.Dbname)
	fmt.Println(psqlInfo)
	gar, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		fmt.Println("Can't connect to db")
	}
	foo1 := new(GarStats) // or &Foo{}
	err = getJson("https://fias.nalog.ru/WebServices/Public/GetLastDownloadFileInfo", &foo1)
	if err != nil {
		fmt.Println(err)
	}
	OldTime := GetCurrentVersion(gar)
	NewTime := ToTimeStamp(foo1.Date)

	if NewTime.After(OldTime) {
		fmt.Println("There is new version...")
		err = downloadFile(foo1.GarXMLFullURL)
		fmt.Println("loading...")
		if err != nil {
			fmt.Println(err)
		}
		InsertData(time.Now(), NewTime, foo1.Version, gar)
	} else {
		fmt.Println("We didn't find new version...")
	}
}
func InsertData(dateDownload, dataRelease time.Time, version int, gar *sqlx.DB) error {
	versionString := strconv.Itoa(version)
	fmt.Println(versionString)
	sqlStatement := "call gar.write_version ('" + dateDownload.String()[0:10] + "', '" + versionString + "');"
	_, err := gar.Exec(sqlStatement)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
func ToTimeStamp(oldTime string) time.Time {
	oldTime = strings.Replace(oldTime, ".", "", -1)
	oldTime = oldTime[4:8] + "-" + oldTime[3:5] + "-" + oldTime[0:2]
	layout := "2006-01-02"
	t, _ := time.Parse(layout, oldTime)
	return t
}
func GetCurrentVersion(gar *sqlx.DB) time.Time {

	var date []string
	err := gar.Select(&date, "select * from gar.get_date_zip()")
	if err != nil {
		return time.Time{}
	}
	date[0] = date[0][0:10]
	layout := "2006-01-02"
	t, _ := time.Parse(layout, date[0])
	return t
}
func downloadFile(url string) (err error) {
	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", url)

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)
	return
}
func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
