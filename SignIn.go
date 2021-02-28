package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/guonaihong/gout"
	"log"
	"time"

	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/wumansgy/goEncrypt"
)

const (
	Skey    = "" //酷推 skey
	Cookie  = "sajssdk_2015_cross_new_user=1; udata_account_300011972819=O7MbhLvKgHdUvHs%2FFo1eTdt0g8kn8IxPDIEALNwCPts%3D; userid_300011972819=1614541679332793714; udata_s_300011972819=1614541680262685408; CAIYUN-TOKEN=jxvPIXEbPaAZCKP/VW6a2sbansrAvW8bo6lYzxIgrEfm18QkPQwVnivg/X+TWsbiH3/b8iwJqlxo7+TckPWx3F2qZqoJYNLByPFSg5KuhCe/i2c+gy1OLJvF7Q69ouDvcOJT7X7YXylWsMHlARfAhKPBl1XWSrbjC3pZmI3ppec8vGYH/nGJm9GExCZS+0dmPjhiy5zvCMwd0OQcIgyPvzwvymJdHu6pJXgHaHOFgcFW8hzhN/Z8rM2AAMX0I6IzQKdLkqXk0l3921cT5mjclA==; CAIYUN-ACCOUNT=NZ2QaAJXzFXkeu7xSME7NA==; CAIYUN-ENCRYPT-ACCOUNT=MTM5Nzg3NDc0OTI=; CAIYUN-SIMPLIFY-ACCOUNT=139****7492; sensorsdata2015jssdkcross=%7B%22distinct_id%22%3A%22177ea2ee7b91c6-0c80114184f5c2-43095b69-230400-177ea2ee7ba297%22%2C%22first_id%22%3A%22%22%2C%22props%22%3A%7B%22%24latest_traffic_source_type%22%3A%22%E7%9B%B4%E6%8E%A5%E6%B5%81%E9%87%8F%22%2C%22%24latest_search_keyword%22%3A%22%E6%9C%AA%E5%8F%96%E5%88%B0%E5%80%BC_%E7%9B%B4%E6%8E%A5%E6%89%93%E5%BC%80%22%2C%22%24latest_referrer%22%3A%22%22%2C%22phoneNumber%22%3A%2213978747492%22%7D%2C%22%24device_id%22%3A%22177ea2ee7b91c6-0c80114184f5c2-43095b69-230400-177ea2ee7ba297%22%7D; sensors_stay_url=https%3A%2F%2Fyun.139.com%2Fm%2F%23%2Fdir%2F1811e671m18g056202102021346567un; sensors_stay_time=1614541879734" //抓包Cookie
	Referer = "https://yun.139.com/m/" //抓包referer
	UA      = "Mozilla/5.0 (Linux; Android 10; M2007J3SC Build/QKQ1.191222.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/83.0.4103.106 Mobile Safari/537.36 MCloudApp/7.6.0"
)

func push(content string) error {
	var resp SendResult
	err := gout.POST(fmt.Sprintf("https://push.xuthus.cc/send/%s", Skey)).
		SetBody(content).Debug(true).BindJSON(&resp).Do()

	if err != nil {
		log.Printf("push err: %v", err)
		return err
	}

	if resp.Code != 200 {
		return errors.New(resp.Message)
	}

	return nil
}

func getEncryptTime() (int64, error) {
	var resp GetEncryptTime
	err := gout.POST("http://caiyun.feixin.10086.cn:7070/portal/ajax/tools/opRequest.action").
		SetHeader(gout.H{
			"Host":             "caiyun.feixin.10086.cn:7070",
			"Accept":           "*/*",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       UA,
			"Content-Type":     "application/x-www-form-urlencoded",
			"Origin":           "http://caiyun.feixin.10086.cn:7070",
			"Referer":          Referer,
			"Accept-Encoding":  "gzip, deflate",
			"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8",
			"Cookie":           Cookie,
		}).Debug(true).SetWWWForm(gout.H{
		"op": "currentTimeMillis",
	}).BindJSON(&resp).Do()
	if err != nil {
		log.Printf("err: %v\n", err)
		return 0, errors.New(err.Error())
	}

	if resp.Code != 10000 {
		log.Printf("err: %v\n", resp.Msg)
		return 0, errors.New(resp.Msg)
	}

	return resp.Result, nil
}

func encryptForm() string {
	t, err := getEncryptTime()
	if err != nil {
		panic(err)
	}

	ef, err := json.Marshal(&EncryptForm{
		SourceId:    1003,
		Type:        1,
		EncryptTime: t,
	})
	if err != nil {
		panic(err)
	}

	var encode = RSAEncrypt(ef)

	return base64.StdEncoding.EncodeToString(encode)
}

func signIn() (*SignInResponse, error) {
	var resp SignInResponse
	err := gout.POST("http://caiyun.feixin.10086.cn:7070/portal/ajax/common/caiYunSignIn.action").
		SetHeader(gout.H{
			"Host":             "caiyun.feixin.10086.cn:7070",
			"Accept":           "*/*",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       UA,
			"Content-Type":     "application/x-www-form-urlencoded",
			"Origin":           "http://caiyun.feixin.10086.cn:7070",
			"Referer":          Referer,
			"Accept-Encoding":  "gzip, deflate",
			"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8",
			"Cookie":           Cookie,
		}).Debug(true).SetWWWForm(gout.H{
		"op":   "receive",
		"data": encryptForm(),
	}).BindJSON(&resp).Do()
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil, errors.New(err.Error())
	}

	if resp.Code != 10000 {
		log.Printf("err: %v\n", resp.Msg)
		return nil, errors.New(resp.Msg)
	}

	return &resp, err
}

func RSAEncrypt(plainText []byte) []byte {
	var publicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCJ6kiv4v8ZcbDiMmyTKvGzxoPR3fTLj/uRuu6dUypy6zDW+EerThAYON172YigluzKslU1PD9+PzPPHLU/cv81q6KYdT+B5w29hlKkk5tNR0PcCAM/aRUQZu9abnl2aAFQow576BRvIS460urnju+Bu1ZtV+oFM+yQu04OSnmOpwIDAQAB
-----END PUBLIC KEY-----`)
	//对明文进行加密
	cipherText, err := goEncrypt.RsaEncrypt(plainText, publicKey)
	if err != nil {
		panic(err)
	}
	//返回密文
	return cipherText
}

func run() (string, error) {
	fmt.Println(time.Now().String(), " 任务执行开始!")

	var content string
	resp, err := signIn()

	if err != nil {
		log.Printf("签到失败: %v", err)
		content = "今日和彩云签到情况: " + err.Error()
		goto Push
	}

	if resp.Result.TodaySignIn {
		content = fmt.Sprintf("和彩云签到情况: 成功\n月签到天数: %d\n总积分: %d",
			resp.Result.MonthDays, resp.Result.TotalPoints)
	}

Push:
	if err = push(content); err != nil {
		log.Println("签到结果: ", content)
	} else {
		log.Println("ok")
	}

	return time.Now().String() + "任务处理完毕！", nil
}

func main() {
	cloudfunction.Start(run)
}
