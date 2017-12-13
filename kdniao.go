package kdniao

import (
	"net/http"
	"net/url"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"fmt"
	"errors"
	"github.com/ridewindx/mel"
	"github.com/ridewindx/mel/binding"
	"github.com/ridewindx/melware"
	"go.uber.org/zap"
	"github.com/tomasen/realip"
)

const (
	BASE_URL      = "http://api.kdniao.cc/api"
	TEST_BASE_URL = "http://testapi.kdniao.cc:8081/api"
)

type Client struct {
	EBusinessID string
	AppKey      string
	*http.Client
}

func NewClient(eid, appKey string, httpClient ...*http.Client) *Client {
	var c *http.Client
	if len(httpClient) > 0 {
		c = httpClient[0]
	} else {
		c = http.DefaultClient
	}
	return &Client{
		EBusinessID: eid,
		AppKey:      appKey,
		Client:      c,
	}
}

type Receiver struct {
	Company      string `json:"Company,omitempty"` // 收件人公司
	Name         string `json:"Name"`              // 收件人
	Tel          string `json:"Tel,omitempty"`     // 电话与手机，必填一个
	Mobile       string `json:"Mobile,omitempty"`
	PostCode     string `json:"PostCode,omitempty"`    // 收件人邮编
	ProvinceName string `json:"ProvinceName"`          // 收件省（如广东省，不要缺少“省”）
	CityName     string `json:"CityName"`              // 收件市（如深圳市，不要缺少“市”）
	ExpAreaName  string `json:"ExpAreaName,omitempty"` // 收件区（如福田区，不要缺少“区”或“县”）
	Address      string `json:"Address"`               // 收件人详细地址
}

type Sender struct {
	Company      string `json:"Company,omitempty"` // 发件人公司
	Name         string `json:"Name"`              // 发件人
	Tel          string `json:"Tel,omitempty"`     // 电话与手机，必填一个
	Mobile       string `json:"Mobile,omitempty"`
	PostCode     string `json:"PostCode,omitempty"`    //	发件人邮编
	ProvinceName string `json:"ProvinceName"`          // 发件省（如广东省，不要缺少“省”）
	CityName     string `json:"CityName"`              // 发件市（如深圳市，不要缺少“市”）
	ExpAreaName  string `json:"ExpAreaName,omitempty"` // 发件区（如福田区，不要缺少“区”或“县”）
	Address      string `json:"Address"`               // 发件详细地址
}

type AddService struct {
	Name       string `json:"Name,omitempty"`       // 增值服务名称
	Value      string `json:"Value,omitempty"`      // 增值服务值
	CustomerID string `json:"CustomerID,omitempty"` // 客户标识(选填)
}

type Commodity struct {
	GoodsName     string  `json:"GoodsName"`               // 商品名称
	GoodsCode     string  `json:"GoodsCode"`     // 商品编码
	Goodsquantity string     `json:"Goodsquantity"` // 件数
	GoodsPrice    string `json:"GoodsPrice"`    // 商品价格
	GoodsWeight   string `json:"GoodsWeight"`   // 商品重量kg
	GoodsDesc     string  `json:"GoodsDesc"`     // 商品描述
	GoodsVol      string `json:"GoodsVol"`      // 商品体积m3
}

type EOrderReq struct {
	CallBack              string      `json:"Callback"`     // 用户自定义回调信息
	MemberID              string      `json:"MemberID"`     // 会员标识，平台方与快递鸟统一用户标识的商家ID
	CustomerName          string      `json:"CustomerName"` // 电子面单客户账号，（与快递网点申请或通过快递鸟官网申请或通过申请电子面单客户号申请）
	CustomerPwd           string      `json:"CustomerPwd"`  // 电子面单密码
	SendSite              string      `json:"SendSite"`     // 收件网点标识
	ShipperCode           string      `json:"ShipperCode"`            // 快递公司编码
	LogisticCode          string      `json:"LogisticCode"` // 快递单号
	ThrOrderCode          string      `json:"ThrOrderCode,omitempty"` // 第三方订单号
	OrderCode             string      `json:"OrderCode"`              // 订单编号
	MonthCode             string      `json:"MonthCode"`    // 月结编码
	PayType               string      `json:"PayType"`                // 邮费支付方式: 1-现付，2-到付，3-月结，4-第三方支付
	ExpType               string      `json:"ExpType"`                // 快递类型：1-标准快件
	IsNotice              string      `json:"IsNotice"`     // 是否通知快递员上门揽件：0-通知；1-不通知；不填则默认为1
	Cost                  string     `json:"Cost"`         // 寄件费（运费）
	OtherCost             string     `json:"OtherCost"`    // 其他费用
	Receiver              Receiver    `json:"Receiver"`
	Sender                Sender      `json:"Sender"`
	StartDate             string      `json:"StartDate"` // 上门取货时间段: "yyyy-MM-dd HH:mm:ss"格式化，本文中所有时间格式相同
	EndDate               string      `json:"EndDate"`
	Weight                string     `json:"Weight"`   // 物品总重量kg
	Quantity              string         `json:"Quantity"` // 件数/包裹数
	Volume                string      `json:"Volume"`   // 物品总体积m3
	Remark                string      `json:"Remark"`   // 备注
	AddService            *AddService  `json:"AddService,omitempty"`
	Commodity             []Commodity `json:"Commodity"`
	IsReturnPrintTemplate string      `json:"IsReturnPrintTemplate"` // 返回电子面单模板：0-不需要；1-需要
	IsSendMessage         string         `json:"IsSendMessage,omitempty"`         // 是否订阅短信：0-不需要；1-需要
	TemplateSize          string      `json:"TemplateSize,omitempty"`          // 模板尺寸
}

type EOrderRep struct {
	EBusinessID string `json:"EBusinessID"` // 电商用户ID
	Order struct {
		OrderCode       string `json:"OrderCode"`                 // 订单编号
		ShipperCode     string `json:"ShipperCode"`               // 快递公司编码
		LogisticCode    string `json:"LogisticCode"`              // 快递单号
		MarkDestination string `json:"MarkDestination,omitempty"` // 大头笔
		OriginCode      string `json:"OriginCode,omitempty"`      // 始发地区域编码
		OriginName      string `json:"OriginName,omitempty"`      // 始发地/始发网点
		DestinatioCode  string `json:"DestinatioCode,omitempty"`  // 目的地区域编码
		DestinatioName  string `json:"DestinatioName,omitempty"`  // 目的地/到达网点
		SortingCode     string `json:"SortingCode,omitempty"`     // 分拣编码
		PackageCode     string `json:"PackageCode,omitempty"`     // 集包编码
	} `json:"Order"`
	Success               bool   `json:"Success"`                         // 成功与否
	ResultCode            string `json:"ResultCode"`                      // 错误编码
	Reason                string `json:"Reason,omitempty"`                // 失败原因
	UniquerRequestNumber  string `json:"UniquerRequestNumber"`            // 唯一标识
	PrintTemplate         string `json:"PrintTemplate,omitempty"`         // 面单打印模板
	EstimatedDeliveryTime string `json:"EstimatedDeliveryTime,omitempty"` // 订单预计到货时间yyyy-mm-dd
	Callback              string `json:"Callback,omitempty"`              // 用户自定义回调信息
	SubCount              int    `json:"SubCount,omitempty"`              // 子单数量
	SubOrders             string `json:"SubOrders,omitempty"`             // 子单号
	SubPrintTemplates     string `json:"SubPrintTemplates,omitempty"`     // 子单模板
	ReceiverSafePhone     string `json:"ReceiverSafePhone,omitempty"`     // 收件人安全电话
	SenderSafePhone       string `json:"SenderSafePhone,omitempty"`       // 寄件人安全电话
	DialPage              string `json:"DialPage"`                        // 拨号页面网址（转换成二维码可扫描拨号）
}

func (c *Client) CreateEOrder(order *EOrderReq) (*EOrderRep, error) {
	order.ExpType = ExpTypeStandard

	data, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	req := c.makeReq(ReqCreateEOder, string(data))

	fmt.Printf("request: %s\n", req.Encode())

	var result EOrderRep
	err = c.post("/eorderservice", req, &result)
	if err != nil {
		return nil, err
	}
	if result.ResultCode != ErrSuccess {
		return nil, fmt.Errorf("code: %s, reason: %s", result.ResultCode, result.Reason)
	}
	return &result, nil
}

type SubscribeTracingReq struct {
	ShipperCode  string `json:"ShipperCode"`  // 快递公司编码
	LogisticCode string `json:"LogisticCode"` // 快递单号
}

func (c *Client) SubscribeTracing(sub *SubscribeTracingReq) error {
	data, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	req := c.makeReq(ReqSubscribeTracing, string(data))

	var rep struct {
		EBusinessID           string `json:"EBusinessID"`
		UpdateTime            string `json:"UpdateTime "`
		Success               bool   `json:"Success"`
		Reason                string `json:"Reason"`
		EstimatedDeliveryTime string `json:"EstimatedDeliveryTime"`
	}

	c.post("/dist", req, &rep)
	if !rep.Success {
		return errors.New(rep.Reason)
	}
	return nil
}

func (c *Client) makeReq(reqType, reqData string) url.Values {
	vals := make(url.Values)
	vals.Set("EBusinessID", c.EBusinessID)
	vals.Set("RequestType", reqType)
	vals.Set("DataSign", c.dataSign(reqData))
	vals.Set("RequestData", reqData)
	vals.Set("DataType", "2")
	return vals
}

func (c *Client) dataSign(data string) string {
	m := md5.New()
	// io.WriteString(m, string(data))
	// io.WriteString(m, c.AppKey)
	data = data + c.AppKey
	fmt.Printf("sign: %s\n", data)
	return base64.StdEncoding.EncodeToString(m.Sum([]byte(data)))
}

func (c *Client) post(relativeURL string, req url.Values, rep interface{}) error {
	r, err := c.Post(BASE_URL+relativeURL, "application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(req.Encode()))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(&rep)
}

type TracingItem struct {
	AcceptTime    string `json:"AcceptTime"`       // 时间
	AcceptStation string `json:"AcceptStation"`    // 描述
	Remark        string `json:"Remark,omitempty"` // 备注
}

type TracingData struct {
	EBusinessID           string        `json:"EBusinessID,omitemtpy"`           // 电商用户ID
	OrderCode             string        `json:"OrderCode,omitempty"`             // 订单编号
	ShipperCode           string        `json:"ShipperCode"`                     // 快递公司编码
	LogisticCode          string        `json:"LogisticCode"`                    // 物流运单号
	Success               bool          `json:"Success"`                         // 成功与否
	Reason                string        `json:"Reason,omitempty"`                // 失败原因
	State                 string        `json:"State"`                           // 物流状态: 0-无轨迹 1-已揽收 2-在途中 3-签收 4-问题件
	CallBack              string        `json:"CallBack,omitempty"`              // 订阅接口的Bk值
	Traces                []TracingItem `json:"Traces"`                          // 物流轨迹详情
	EstimatedDeliveryTime string        `json:"EstimatedDeliveryTime,omitempty"` // 预计到达时间yyyy-mm-dd
	PickerInfo struct {
		PersonName     string `json:"PersonName,omitempty"`     // 快递员姓名
		PersonTel      string `json:"PersonTel,omitempty"`      // 快递员电话
		PersonCode     string `json:"PersonCode,omitempty"`     // 快递员工号
		StationName    string `json:"StationName,omitempty"`    // 网点名称
		StationAddress string `json:"StationAddress,omitempty"` // 网点地址
		StationTel     string `json:"StationTel,omitempty"`     // 网点电话
	} `json:"PickerInfo,omitempty"`                                              // 收件员信息
	SenderInfo struct {
		PersonName     string `json:"PersonName,omitempty"`     // 派件员姓名
		PersonTel      string `json:"PersonTel,omitempty"`      // 派件员快递员电话
		PersonCode     string `json:"PersonCode,omitempty"`     // 派件员快递员工号
		StationName    string `json:"StationName,omitempty"`    // 派件员网点名称
		StationAddress string `json:"StationAddress,omitempty"` // 派件员网点地址
		StationTel     string `json:"StationTel,omitempty"`     // 派件员网点电话
	} `json:"SenderInfo,omitempty"`                                              // 派件员信息
}

func PushHandler(c *mel.Context, tracingHandler func([]TracingData)) {
	var req struct {
		RequestType string
		DataSign    string
		RequestData string
	}
	err := c.BindWith(&req, binding.FormPost)
	if err != nil {
		c.AbortWithError(400, err).Type = mel.ErrorTypeBind
		return
	}

	switch req.RequestType {
	case PushTracing:
		data, err := url.QueryUnescape(req.RequestData)
		var tracing struct {
			EBusinessID string        `json:"EBusinessID"`
			PushTime    string        `json:"PushTime"`
			Count       int           `json:"Count"`
			Data        []TracingData `json:"Data"`
		}
		if err == nil {
			err = json.Unmarshal([]byte(data), &tracing)
		}
		if err != nil {
			c.AbortWithError(400, err).Type = mel.ErrorTypePrivate
			return
		}
		tracingHandler(tracing.Data)
		rep := struct {
			EBusinessID string `json:"EBusinessID"`
			UpdateTime  string `json:"UpdateTime"`
			Success     bool   `json:"Success"`
		}{
			EBusinessID: tracing.EBusinessID,
			UpdateTime:  tracing.PushTime,
			Success:     true,
		}
		c.JSON(200, &rep)
	}
}

func PrintHandler(c *mel.Context, client *Client) {
	var req []struct {
		OrderID   string `json:"OrderCode"`
		PrintName string `json:"PortName"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(400, err).Type = mel.ErrorTypeBind
		return
	}
	data, err := json.Marshal(req)
	if err != nil {
		c.AbortWithError(400, err).Type = mel.ErrorTypeBind
		return
	}

	sign := client.dataSign(realip.RealIP(c.Request) + url.QueryEscape(string(data)))

	rep := struct {
		EID       string `json:"eid"`
		Signature string `json:"signature"`
	}{
		EID:       client.EBusinessID,
		Signature: sign,
	}
	c.JSON(200, &rep)
}

type Server struct {
	*mel.Mel
}

func NewServer(pushURL string, tracingHandler func([]TracingData), logger *zap.SugaredLogger) *Server {
	s := &Server{
		Mel: mel.New(),
	}

	s.Use(melware.Zap(logger))

	s.Post(pushURL, func(c *mel.Context) {
		PushHandler(c, tracingHandler)
	})

	return s
}

func (s *Server) HandlePrint(printURL string, client *Client) {
	s.Post(printURL, func(c *mel.Context) {
		PrintHandler(c, client)
	})
}
