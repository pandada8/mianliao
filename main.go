package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
	"time"
	// "terminal"
)

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (self *Config) loadConfig(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, self)
}

func (self *Config) saveConfig(filename string) error {
	content, err := json.Marshal(*self)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, content, 0644)
}

func (self *Config) promptConfig() error {
	var err error
	fmt.Print("Username: ")
	// self.username = terminal.ReadPassword()
	_, err = fmt.Scanf("%s", &self.Username)
	if err != nil {
		return err
	}
	fmt.Print("Password: ")
	_, err = fmt.Scanf("%s", &self.Password)
	return err
}

func sayHello() {
	fmt.Println(`
   _____  .__              .__  .__
  /     \ |__|____    ____ |  | |__|____    ____
 /  \ /  \|  \__  \  /    \|  | |  \__  \  /  _ \
/    Y    \  |/ __ \|   |  \  |_|  |/ __ \(  <_> )
\____|__  /__(____  /___|  /____/__(____  /\____/
        \/        \/     \/             \/
`)
}

func showContent(ret *http.Response) {
	content, err := ioutil.ReadAll((*ret).Body)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(content))
	}
	defer (*ret).Body.Close()
}

func (self *Config) login() error {
	jar, _ := cookiejar.New(nil)

	// the mianliao using a self-signed cert, just skip the cert verify
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	c := &http.Client{Jar: jar, Transport: tr}

	info := make(url.Values)
	info.Add("ua", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36")
	info.Add("sw", "1280")
	info.Add("sh", "720")
	info.Add("ww", "1280")
	info.Add("wh", "720")

	data := make(url.Values)
	data.Add("action", "login")
	data.Add("username", self.Username)
	data.Add("password", self.Password)

	ret, err := c.PostForm("https://wifi.52mianliao.com", info)
	// showContent(ret)
	ret, err = c.PostForm("https://wifi.52mianliao.com", data)
	// showContent(ret)
	ret, err = c.PostForm("https://wifi.52mianliao.com", info)
	if err != nil {
		return err
	} else {
		if content, err := ioutil.ReadAll(ret.Body); err != nil {
			return err
		} else {
			result := string(content)
			if strings.Contains(result, "登陆用户") {
				re, _ := regexp.Compile("<label>(.*)</label>")
				found := re.FindAllStringSubmatch(result, -1)
				for i := 0; i < len(found); i++ {
					fmt.Println(found[i][1])
				}
			} else {
				return errors.New("Login failed")
			}
		}
	}
	return nil
}

func main() {
	sayHello()
	var cfg Config
	usr, _ := user.Current()
	home := usr.HomeDir
	cfgPath := path.Join(home, "mianliao.json")
	var err error

	if _, err = os.Stat(cfgPath); os.IsNotExist(err) {
		err = cfg.promptConfig()
	} else {
		fmt.Printf("Try loading infomation from: %s\n", cfgPath)
		err = cfg.loadConfig(cfgPath)
		if len(cfg.Password) == 0 {
			cfg.promptConfig()
		}
	}
	if err != nil {
		fmt.Println(err)
	}

	if err = cfg.login(); err != nil {
		fmt.Println(err)
	} else {
		if err = cfg.saveConfig(cfgPath); err != nil {
			fmt.Println(err)
		}
	}
	time.Sleep(2 * time.Second)
}
