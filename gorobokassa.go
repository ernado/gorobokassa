package gorobokassa

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	queryOutSumm     = "OutSum"
	queryInvID       = "InvId"
	queryCRC         = "SignatureValue"
	queryDescription = "Desc"
	queryLogin       = "MrchLogin"
	robokassaHost    = "auth.robokassa.ru"
	robokassaPath    = "Merchant/Index.aspx"
	scheme           = "https"
	delim            = ":"
)

// Client для генерации URL и проверки уведомлений
type Client struct {
	login          string
	firstPassword  string
	secondPassword string
}

// URL переадресации пользователя на оплату
func (client *Client) URL(invoice, value int, description string) string {
	return buildRedirectURL(client.login, client.firstPassword, invoice, value, description)
}

// CheckResult получение уведомления об исполнении операции (ResultURL)
func (client *Client) CheckResult(r *http.Request) bool {
	return verifyRequest(client.secondPassword, r)
}

// CheckSuccess проверка параметров в скрипте завершения операции (SuccessURL)
func (client *Client) CheckSuccess(r *http.Request) bool {
	return verifyRequest(client.firstPassword, r)
}

// New Client
func New(login, password1, password2 string) *Client {
	return &Client{login, password1, password2}
}

// CRC of joint values with delimeter
func CRC(v ...interface{}) string {
	s := make([]string, len(v))
	for key, value := range v {
		s[key] = fmt.Sprintf("%v", value)
	}
	h := md5.New()
	io.WriteString(h, strings.Join(s, delim))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func buildRedirectURL(login, password string, invoice, value int, description string) string {
	q := url.URL{}
	q.Host = robokassaHost
	q.Scheme = scheme
	q.Path = robokassaPath

	params := url.Values{}
	params.Add(queryLogin, login)
	params.Add(queryOutSumm, strconv.Itoa(value))
	params.Add(queryInvID, strconv.Itoa(invoice))
	params.Add(queryDescription, description)
	params.Add(queryCRC, CRC(login, value, invoice, password))

	q.RawQuery = params.Encode()
	return q.String()
}

func verifyResult(password string, invoice, value int, crc string) bool {
	return strings.ToUpper(crc) == strings.ToUpper(CRC(value, invoice, password))
}

func verifyRequest(password string, r *http.Request) bool {
	q := r.URL.Query()
	value, err := strconv.Atoi(q.Get(queryOutSumm))
	if err != nil {
		log.Println(err)
		return false
	}
	invoice, err := strconv.Atoi(q.Get(queryInvID))
	if err != nil {
		log.Println(err)
		return false
	}
	crc := q.Get(queryCRC)
	return verifyResult(password, invoice, value, crc)
}
