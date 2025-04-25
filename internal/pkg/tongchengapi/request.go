package tongchengapi

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"myGin/internal/pkg/tcrypt" // 更新了导入路径

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

func Test(option Options) {
	// 测试代码
	option.Set("DeviceId", "433953b3f3805651")
	// fmt.Println("DeviceId:", option.Get("DeviceId"))
	// fmt.Println("boolean:", option.Get("boolean"))
	// fmt.Println("int:", option.Get("int"))
}

func get_smstoken()(string , error){
	client := req.C()
	resp, err := client.R().
    SetHeader("User-Agent", "My-Custom-Client").
    Get("http://127.0.0.1:8787/api/sm")
	if err != nil {
		fmt.Println("请求失败:", err)
		return "请求失败",err
	}
	// fmt.Println("get_smstoken response text:", resp.String())
	rid := gjson.Get(resp.String(), "rid").String()
	return rid,nil
}

func SendSms(options Options) (string, error) {
	smstoken,error := get_smstoken()
	if error != nil {
		fmt.Println("获取 smstoken 失败:", error)
		return "", error
	}

	options.Set("smtoken", smstoken)
	// fmt.Println("smtoken:", options.Get("smtoken"))
	// fmt.Println("mobile:", options.Get("mobile"))



	client := req.C()
	
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	digitalSign := tcrypt.Md5Encrypt("AccountID=c26b007f-c89e-431a-b8cc-493becbdd8a2&ReqTime=" + timeStr + "&ServiceName=sendsms&Version=201111281029128874d8a8b8b391fbbd1a25bda6ecda11")

    requestBody := map[string]interface{}{
		"request": map[string]interface{}{
			"body": map[string]interface{}{
				"mobile": options.Get("mobile"),
				"action": "loginHalf",
				"clientInfo": map[string]interface{}{
					"versionType": "android",
					"area": "||",
					"clientId": "8a5a12e00459575439e34a0118c598af972f55cda110",
					"hwPreInstall": "",
					"isGuest": "false",
					"deviceId": options.Get("deviceId"),
					"versionNumber": "11.0.6",
					"touristId": "e390d4da0593404c9ea828f533d516b7",
					"manufacturer": "Xiaomi",
					"extend": "4^13,5^22041216C,isGuest^false,6^-1,os_v^33,app_v^11.0.6.1,devicetoken^,tencentDeviceToken^v2:zqkBrv8m011woL+o+R3ezRUVcqwA5H1uqtxIfUl1vblBjw6omkg2AhESIaU+lahNdbxp6PgX9ge5ZBTbkX/Q1VpQAUxAwT0yCmNJIkOXyEqnR4vg7MgY2u0RiveJGePB1oeTlEqEBR62mQENS74GcO1FOUW10LLSrWX4PMC+xJDe9l7baQVFuhDt2QMIt/+f4tBInBwonIVC9S4sERexjXxIYDylTPnkFlIqNgo8X2iTV3WDoJAXDuowsFHVhTwgagSifrKavVa5irVdrZ7BbcjSXDpX1YtoiBWRbdeafjeIuqTPEchv1iysQg==",
					"systemCode": "tc",
					"clientIp": "127.0.0.1",
					"refId": "16359978",
					"networkType": "wifi",
					"device": fmt.Sprintf("%s|arm64-v8a|1080*2316*440|22041216C|unknown",options.Get("deviceId")),
					"pushInfo": "QFU8dDEUBWnHlVW0fFCByPtqyF5OtG+rG7491UYf2sPg4/wb4AZ+ikBE4F6joUnm",
				},
				"areaCode": "86",
				"type": "sm",
				"token": options.Get("smtoken"),
			},
			"header": map[string]interface{}{
				"accountID": "c26b007f-c89e-431a-b8cc-493becbdd8a2",
				"digitalSign": digitalSign,
				"reqTime": timeStr,
				"serviceName": "sendsms",
				"version": "20111128102912",
			},
		},
	}
    
    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf("JSON序列化错误: %v", err)
    }
	jsonstr := string(jsonData)
	// fmt.Println("jsonstring:", jsonstr)

	body_text , _ := tcrypt.AesEcbEncrypt(jsonstr)




	// 请求体
	body := `{"sv":2,"data":"`+body_text+`","key":"Sg7FNdwhg3EApfUqiv0tdqhEhIdZE/02G/OMPFdrYGpHjZh7LUuzJIZyVqGuBDUuZTCm7MSCwo29cMCaZ6WBEKTfmFZUUDUusyPkITTvKxz5A1OH3qimML/hPHcbgRPPieMX8MfFsJxEUJYkdqEyK9jAHo5s3R4ZVhkyJbk1bjg31HKJgK4NtBmcQxRUmSvaHwKvOtFm12XwEvoQCX7Qd4yjmRtaYNCdjxdUWgRM2fvMo3ab4OqgYCfK8ruQlyqT+A+ZUPIhdlSB3z5tIv2wmai5d6wZYoEtW+rI9d05SflmSxRp8vg+iu+HkrPvwQGXG9qow/9pOvyibBxZpDw5ww=="}`

	// fmt.Println("请求体:", body)

	reqdata := tcrypt.Md5Encrypt(jsonstr + "4957CA66-37C3-46CB-B26D-E3D9DCB51535")

	// fmt.Println("reqdata:", reqdata)

	// 设置请求头
	headers := map[string]string{
		"User-Agent":     "okhttp/3.12.13",
		"Connection":     "Keep-Alive",
		"Accept-Encoding": "gzip",
		"Content-Type":   "application/json; charset=utf-8",
		"reqdata":       reqdata,
	}
	// fmt.Println("请求头:", headers)
	// 发送 POST 请求
	resp, err := client.
		SetProxyURL("http://127.0.0.1:9000").
		R().
		SetHeaders(headers).
		SetBody(body).
		Post("https://tcmobileapi.17usoft.com/member/MemberHandler.ashx")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	// fmt.Println("响应状态码:", resp.StatusCode)
	// fmt.Println("sendsms:", resp.String())
	rspcode := gjson.Get(resp.String(), "response.header.rspCode").String()
	if rspcode != "0000" {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("请求失败: %s", rspcode)
	} else {
		fmt.Println("请求成功:", resp.String())
	}
	return rspcode, nil
}

func Check_smscode(options Options) (string, error) {

	client := req.C()
	
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	digitalSign := tcrypt.Md5Encrypt("AccountID=c26b007f-c89e-431a-b8cc-493becbdd8a2&ReqTime=" + timeStr + "&ServiceName=checksms&Version=201111281029128874d8a8b8b391fbbd1a25bda6ecda11")

    requestBody := map[string]interface{}{
		"request": map[string]interface{}{
			"body": map[string]interface{}{
				"mobile": options.Get("mobile"),
				"action": "loginHalf",
				"clientInfo": map[string]interface{}{
					"versionType": "android",
					"area": "||",
					"clientId": "8a5a12e00459575439e34a0118c598af972f55cda110",
					"hwPreInstall": "",
					"isGuest": "false",
					"deviceId": options.Get("deviceId"),
					"versionNumber": "11.0.6",
					"touristId": "e390d4da0593404c9ea828f533d516b7",
					"manufacturer": "Xiaomi",
					"extend": "4^13,5^22041216C,isGuest^false,6^-1,os_v^33,app_v^11.0.6.1,devicetoken^,tencentDeviceToken^v2:zqkBrv8m011woL+o+R3ezRUVcqwA5H1uqtxIfUl1vblBjw6omkg2AhESIaU+lahNdbxp6PgX9ge5ZBTbkX/Q1VpQAUxAwT0yCmNJIkOXyEqnR4vg7MgY2u0RiveJGePB1oeTlEqEBR62mQENS74GcO1FOUW10LLSrWX4PMC+xJDe9l7baQVFuhDt2QMIt/+f4tBInBwonIVC9S4sERexjXxIYDylTPnkFlIqNgo8X2iTV3WDoJAXDuowsFHVhTwgagSifrKavVa5irVdrZ7BbcjSXDpX1YtoiBWRbdeafjeIuqTPEchv1iysQg==",
					"systemCode": "tc",
					"clientIp": "127.0.0.1",
					"refId": "16359978",
					"networkType": "wifi",
					"device": fmt.Sprintf("%s|arm64-v8a|1080*2316*440|22041216C|unknown",options.Get("deviceId")),
					"pushInfo": "QFU8dDEUBWnHlVW0fFCByPtqyF5OtG+rG7491UYf2sPg4/wb4AZ+ikBE4F6joUnm",
				},
				"areaCode": "86",
				"verifyCode": options.Get("verifyCode"),
			},
			"header": map[string]interface{}{
				"accountID": "c26b007f-c89e-431a-b8cc-493becbdd8a2",
				"digitalSign": digitalSign,
				"reqTime": timeStr,
				"serviceName": "checksms",
				"version": "20111128102912",
			},
		},
	}
    
    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf("JSON序列化错误: %v", err)
    }
	jsonstr := string(jsonData)
	body_text , _ := tcrypt.AesEcbEncrypt(jsonstr)
	// 请求体
	body := `{"sv":2,"data":"`+body_text+`","key":"Sg7FNdwhg3EApfUqiv0tdqhEhIdZE/02G/OMPFdrYGpHjZh7LUuzJIZyVqGuBDUuZTCm7MSCwo29cMCaZ6WBEKTfmFZUUDUusyPkITTvKxz5A1OH3qimML/hPHcbgRPPieMX8MfFsJxEUJYkdqEyK9jAHo5s3R4ZVhkyJbk1bjg31HKJgK4NtBmcQxRUmSvaHwKvOtFm12XwEvoQCX7Qd4yjmRtaYNCdjxdUWgRM2fvMo3ab4OqgYCfK8ruQlyqT+A+ZUPIhdlSB3z5tIv2wmai5d6wZYoEtW+rI9d05SflmSxRp8vg+iu+HkrPvwQGXG9qow/9pOvyibBxZpDw5ww=="}`

	reqdata := tcrypt.Md5Encrypt(jsonstr + "4957CA66-37C3-46CB-B26D-E3D9DCB51535")

	// 设置请求头
	headers := map[string]string{
		"User-Agent":     "okhttp/3.12.13",
		"Connection":     "Keep-Alive",
		// "Accept-Encoding": "gzip",
		"Content-Type":   "application/json; charset=utf-8",
		"reqdata":       reqdata,
	}
	// fmt.Println("请求头:", headers)
	resp, err := client.
		SetProxyURL("http://127.0.0.1:9000").
		R().
		SetHeaders(headers).
		SetBody(body).
		Post("https://tcmobileapi.17usoft.com/member/MemberHandler.ashx")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	rspcode := gjson.Get(resp.String(), "response.header.rspCode").String()
	// rspdesc := gjson.Get(resp.String(), "response.header.rspDesc").String()
	if rspcode != "0000" {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("请求失败: %s",resp.String())
	} else {
		signKey := gjson.Get(resp.String(), "response.body.sign").String()
		fmt.Println("check_smscode请求成功:", resp.String())
		options.Set("signKey",signKey)
	}
	return rspcode, nil
}
func Login_smscode(options Options) (string, error) {

	client := req.C()
	
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	digitalSign := tcrypt.Md5Encrypt("AccountID=c26b007f-c89e-431a-b8cc-493becbdd8a2&ReqTime=" + timeStr + "&ServiceName=loginbyvalidatecode&Version=201111281029128874d8a8b8b391fbbd1a25bda6ecda11")

    requestBody := map[string]interface{}{
		"request": map[string]interface{}{
			"body": map[string]interface{}{
				"mobile": options.Get("mobile"),
				"clientInfo": map[string]interface{}{
					"versionType": "android",
					"area": "||",
					"clientId": "8a5a12e00459575439e34a0118c598af972f55cda110",
					"hwPreInstall": "",
					"isGuest": "false",
					"deviceId": options.Get("deviceId"),
					"versionNumber": "11.0.6",
					"touristId": "e390d4da0593404c9ea828f533d516b7",
					"manufacturer": "Xiaomi",
					"extend": "4^13,5^22041216C,isGuest^false,6^-1,os_v^33,app_v^11.0.6.1,devicetoken^,tencentDeviceToken^v2:zqkBrv8m011woL+o+R3ezRUVcqwA5H1uqtxIfUl1vblBjw6omkg2AhESIaU+lahNdbxp6PgX9ge5ZBTbkX/Q1VpQAUxAwT0yCmNJIkOXyEqnR4vg7MgY2u0RiveJGePB1oeTlEqEBR62mQENS74GcO1FOUW10LLSrWX4PMC+xJDe9l7baQVFuhDt2QMIt/+f4tBInBwonIVC9S4sERexjXxIYDylTPnkFlIqNgo8X2iTV3WDoJAXDuowsFHVhTwgagSifrKavVa5irVdrZ7BbcjSXDpX1YtoiBWRbdeafjeIuqTPEchv1iysQg==",
					"systemCode": "tc",
					"clientIp": "127.0.0.1",
					"refId": "16359978",
					"networkType": "wifi",
					"device": "433953b3f3805658|arm64-v8a|1080*2316*440|22041216C|unknown",
					"pushInfo": "QFU8dDEUBWnHlVW0fFCByPtqyF5OtG+rG7491UYf2sPg4/wb4AZ+ikBE4F6joUnm",
				},
				"areaCode": "86",
				"signKey": options.Get("signKey"),
				"secInfo": "B6/ScOzk58unyYD35pqGGLJ4O/mizu9YD3ndTBi0AofR+mwM1rB1nWfVrUvC4jyIyk+Jw9t+MdBkCtLunZUsflcDZvpGYlT4O3K83z2FNTW/rl407pejCtPxqF7XP9RDgCeKeB/w67QZgal6AIuLeffI1bS/6IiGT2SqZUdZYYjnfAZPtlAXBQO2RI8qF+nebRqducGKvQZXF92+5yMBdGNXD+oViikURx3pqwXRFmrNam2fhx/Ln0Z5HazeQenWq0Y7Wj+xN2C2nFGUgngJiupBT2mCbxzz5TgStBJ6JPg=",
				"verifyCode": options.Get("verifyCode"),
			},
			"header": map[string]interface{}{
				"accountID": "c26b007f-c89e-431a-b8cc-493becbdd8a2",
				"digitalSign": digitalSign,
				"reqTime": timeStr,
				"serviceName": "loginbyvalidatecode",
				"version": "20111128102912",
			},
		},
	}
    
    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf("JSON序列化错误: %v", err)
    }
	jsonstr := string(jsonData)
	body_text , _ := tcrypt.AesEcbEncrypt(jsonstr)
	// 请求体
	body := `{"sv":2,"data":"`+body_text+`","key":"Sg7FNdwhg3EApfUqiv0tdqhEhIdZE/02G/OMPFdrYGpHjZh7LUuzJIZyVqGuBDUuZTCm7MSCwo29cMCaZ6WBEKTfmFZUUDUusyPkITTvKxz5A1OH3qimML/hPHcbgRPPieMX8MfFsJxEUJYkdqEyK9jAHo5s3R4ZVhkyJbk1bjg31HKJgK4NtBmcQxRUmSvaHwKvOtFm12XwEvoQCX7Qd4yjmRtaYNCdjxdUWgRM2fvMo3ab4OqgYCfK8ruQlyqT+A+ZUPIhdlSB3z5tIv2wmai5d6wZYoEtW+rI9d05SflmSxRp8vg+iu+HkrPvwQGXG9qow/9pOvyibBxZpDw5ww=="}`

	reqdata := tcrypt.Md5Encrypt(jsonstr + "4957CA66-37C3-46CB-B26D-E3D9DCB51535")

	// 设置请求头
	headers := map[string]string{
		"User-Agent":     "okhttp/3.12.13",
		"Connection":     "Keep-Alive",
		// "Accept-Encoding": "gzip",
		"Content-Type":   "application/json; charset=utf-8",
		"reqdata":       reqdata,
	}
	// fmt.Println("请求头:", headers)
	resp, err := client.
		SetProxyURL("http://127.0.0.1:9000").
		R().
		SetHeaders(headers).
		SetBody(body).
		Post("https://tcmobileapi.17usoft.com/member/MemberHandler.ashx")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	// rspcode := gjson.Get(resp.String(), "response.header.rspCode").String()
	rspdesc := gjson.Get(resp.String(), "response.header.rspDesc").String()

	fmt.Println("login_smscode请求成功:", resp.String())
	if rspdesc == "未注册"{
		Regist_smscode(options)
		return rspdesc, fmt.Errorf("login_smscode请求成功: %s",rspdesc)
	}else if rspdesc == "验证码错误"{
		return rspdesc, fmt.Errorf("login_smscode请求成功: %s",rspdesc)
	}
	
	return rspdesc, nil
}
func Regist_smscode(options Options) (string, error) {

	client := req.C()
	
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	digitalSign := tcrypt.Md5Encrypt("AccountID=c26b007f-c89e-431a-b8cc-493becbdd8a2&ReqTime=" + timeStr + "&ServiceName=registerbyvalidatecodeorder&Version=201111281029128874d8a8b8b391fbbd1a25bda6ecda11")

    requestBody := map[string]interface{}{
		"request": map[string]interface{}{
			"body": map[string]interface{}{
				"mobile": "17115180520",
				"clientInfo": map[string]interface{}{
					"versionType": "android",
					"area": "||",
					"clientId": "8a5a12e00459575439e34a0118c598af972f55cda110",
					"hwPreInstall": "",
					"isGuest": "false",
					"deviceId": "433953b3f3805658",
					"versionNumber": "11.0.6",
					"touristId": "e390d4da0593404c9ea828f533d516b7",
					"manufacturer": "Xiaomi",
					"extend": "4^13,5^22041216C,isGuest^false,6^-1,os_v^33,app_v^11.0.6.1,devicetoken^,tencentDeviceToken^v2:zqkBrv8m011woL+o+R3ezRUVcqwA5H1uqtxIfUl1vblBjw6omkg2AhESIaU+lahNdbxp6PgX9ge5ZBTbkX/Q1VpQAUxAwT0yCmNJIkOXyEqnR4vg7MgY2u0RiveJGePB1oeTlEqEBR62mQENS74GcO1FOUW10LLSrWX4PMC+xJDe9l7baQVFuhDt2QMIt/+f4tBInBwonIVC9S4sERexjXxIYDylTPnkFlIqNgo8X2iTV3WDoJAXDuowsFHVhTwgagSifrKavVa5irVdrZ7BbcjSXDpX1YtoiBWRbdeafjeIuqTPEchv1iysQg==",
					"systemCode": "tc",
					"clientIp": "127.0.0.1",
					"refId": "16359978",
					"networkType": "wifi",
					"device": "433953b3f3805658|arm64-v8a|1080*2316*440|22041216C|unknown",
					"pushInfo": "QFU8dDEUBWnHlVW0fFCByPtqyF5OtG+rG7491UYf2sPg4/wb4AZ+ikBE4F6joUnm",
				},
				"areaCode": "86",
				"signKey": options.Get("signKey"),
				"secInfo": "B6/ScOzk58unyYD35pqGGLJ4O/mizu9YD3ndTBi0AofR+mwM1rB1nWfVrUvC4jyIyk+Jw9t+MdBkCtLunZUsflcDZvpGYlT4O3K83z2FNTW/rl407pejCtPxqF7XP9RDgCeKeB/w67QZgal6AIuLeffI1bS/6IiGT2SqZUdZYYjnfAZPtlAXBQO2RI8qF+nebRqducGKvQZXF92+5yMBdGNXD+oViikURx3pqwXRFmrNam2fhx/Ln0Z5HazeQenW8lqyGgytVm9VG121Ie7o3d1QD5KLnuhZMVnDq4nvmxs=",
				"verifyCode": options.Get("verifyCode"),
			},
			"header": map[string]interface{}{
				"accountID": "c26b007f-c89e-431a-b8cc-493becbdd8a2",
				"digitalSign": digitalSign,
				"reqTime": timeStr,
				"serviceName": "registerbyvalidatecodeorder",
				"version": "20111128102912",
			},
		},
	}
    
    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf("JSON序列化错误: %v", err)
    }
	jsonstr := string(jsonData)
	body_text , _ := tcrypt.AesEcbEncrypt(jsonstr)
	// 请求体
	body := `{"sv":2,"data":"`+body_text+`","key":"Sg7FNdwhg3EApfUqiv0tdqhEhIdZE/02G/OMPFdrYGpHjZh7LUuzJIZyVqGuBDUuZTCm7MSCwo29cMCaZ6WBEKTfmFZUUDUusyPkITTvKxz5A1OH3qimML/hPHcbgRPPieMX8MfFsJxEUJYkdqEyK9jAHo5s3R4ZVhkyJbk1bjg31HKJgK4NtBmcQxRUmSvaHwKvOtFm12XwEvoQCX7Qd4yjmRtaYNCdjxdUWgRM2fvMo3ab4OqgYCfK8ruQlyqT+A+ZUPIhdlSB3z5tIv2wmai5d6wZYoEtW+rI9d05SflmSxRp8vg+iu+HkrPvwQGXG9qow/9pOvyibBxZpDw5ww=="}`

	reqdata := tcrypt.Md5Encrypt(jsonstr + "4957CA66-37C3-46CB-B26D-E3D9DCB51535")

	// 设置请求头
	headers := map[string]string{
		"User-Agent":     "okhttp/3.12.13",
		"Connection":     "Keep-Alive",
		"Accept-Encoding": "gzip",
		"Content-Type":   "application/json; charset=utf-8",
		"reqdata":       reqdata,
	}
	// fmt.Println("请求头:", headers)
	resp, err := client.
		SetProxyURL("http://127.0.0.1:9000").
		R().
		SetHeaders(headers).
		SetBody(body).
		Post("https://tcmobileapi.17usoft.com/member/MemberHandler.ashx")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	rspcode := gjson.Get(resp.String(), "response.header.rspCode").String()
	rspdesc := gjson.Get(resp.String(), "response.header.rspDesc").String()

	// fmt.Println("regist_smscode请求成功:", resp.String())
	if rspcode != "0000" {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("请求失败: %s", rspdesc)
	} else {
		fmt.Println("regist_smscode请求成功:", resp.String())
		secToken := gjson.Get(resp.String(), "response.body.securityToken").String()
		tcuserid := gjson.Get(resp.String(), "response.body.memberId").String()
		tcsectoken := gjson.Get(resp.String(), "response.body.externalMemberId").String()
		options.Set("secToken", secToken)
		options.Set("tcuserid", tcuserid)
		options.Set("tcsectoken", tcsectoken)
	}

	
	return rspdesc, nil
}
func Get_airline_message(options Options) (string, error) {

	client := req.C()
	

    requestBody := map[string]interface{}{
		"dcc": "SHA",
		"acc": "BJS",
		"pt": 0,
		"ddate": "2025-05-20",
		"cabin": 0,
		"cc": 0,
		"entrance": 0,
		"rtm": 0,
		"passengerInfo": map[string]interface{}{
			"adultCount": 1,
			"childCount": 0,
			"infantCount": 0,
		},
		"pc": map[string]interface{}{
			"sd": "2025-05-20",
			"ed": "2025-05-20",
		},
	}
    
    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return "", fmt.Errorf("JSON序列化错误: %v", err)
    }
	jsonstr := string(jsonData)


	// reqdata := tools.Md5Encrypt(jsonstr + "4957CA66-37C3-46CB-B26D-E3D9DCB51535")

	// 设置请求头
	headers := map[string]string{
		"Host":            "wx.17u.cn",
		"Connection":      "keep-alive",
		"Content-Length":  "324",
		"User-Agent":      "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Accept":          "application/json, text/plain, */*",
		"Content-Type":    "application/json",
		"tcunionid":       "",
		"tcplat":          "434",
		"tcuserid":        fmt.Sprintf("%s", options.Get("tcuserid")),
		"isLogin":         "true",
		"diysearch":       "1",
		"tcappv":          "11.0.6",
		"cache-control":   "no-cache",
		"tcsectoken":      fmt.Sprintf("%s", options.Get("tcsectoken")),
		"devicetype":      "Xiaomi|22041216C",
		"Origin":          "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site":  "same-origin",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Dest":  "empty",
		// "Referer":         "https://wx.17u.cn/fapp/app/book1?date=2025-03-19&backDate=&childticket=0,0&fromCode=SHA&toCode=PEK&fromCity=%E4%B8%8A%E6%B5%B7&toCity=%E5%8C%97%E4%BA%AC&fromcitycode=SHA&tocitycode=BJS&dcn=%E4%B8%8A%E6%B5%B7&acn=%E5%8C%97%E4%BA%AC&fromNear=0&cabin=0&hasAirPort=0&fromcitytype=0&hasfx=0&nametype=0,0&fromUnionsearch=&refId=&cheapflightid=&defaultAirCompanys=&crossrequestid=&frompage=HOME&an=1&cn=0&baby=0&platcode=10189&direct=0&thirdMemberId=&poikey=&outrefid=&channelid=434&takeChildOrBaby=0&ofromcitycode=SHA&otocitycode=BJS&platCode=10189",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		// "Accept-Encoding": "gzip",
	}
	
	// fmt.Println("请求头:", headers)
	resp, err := client.
		// SetProxyURL("http://127.0.0.1:9000").
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightbffv2/book1/flights")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	resps := gjson.Get(resp.String(), "success").Bool()
	// fmt.Println("regist_smscode请求成功:", resp.String())
	if resps{
		// fmt.Println("getairline-message请求成功:", resp.String())
		sift(options, resp.String())
	}else {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("getairline-message请求失败: %s", resp.String())
	}

	
	return resp.String(), nil
}

func sift(options Options, data string){
	fmt.Println("开始获取航线信息：",options.Get("tcuserid"))
	flights := gjson.Get(data, "data.fl")
	count := flights.Array()
	fmt.Println("获取到的航线条数：",len(count))
	for i:=0; i<len(count);i++{
		jsonPathList :=[]string{"dac","aac","fn","dt","at","dasn","aasn","asn","amn","lps.0.sp", "lps.0.atp"}
		jsonPathNameList :=[]string{"出发机场","到达机场","航班号","起飞时间","到达时间","出发航站楼","到达航站楼","航空公司名称","航空公司简称","经济舱最低价", "航运价"}
		// 创建一个字符串构建器来高效拼接字符串
		var infoBuilder strings.Builder
		for h:=0;h<len(jsonPathList);h++{
			path := fmt.Sprintf("data.fl.%d.%s", i, jsonPathList[h])
			value := gjson.Get(data, path).String()
			// 添加空格，除了第一个字段
			if h>= 0 {
				infoBuilder.WriteString("")
		}
		 // 拼接字段名和值
		infoBuilder.WriteString(fmt.Sprintf("%s: %s\n", jsonPathNameList[h], value))
		// 打印航班编号和所有字段信息
		// fmt.Printf("航班 #%d 信息:%s\n", i+1, infoBuilder.String())
		
	}

	
}
AddPassenger(options)
// ListPassenger(options)
}

func AddPassenger(options Options) (string, error) {
	fmt.Println("mobile:",options.Get("mobile"))
	client := req.C()

	// Construct request body
	requestBody := map[string]interface{}{
        "birthday":     options.Get("passenger_birthday"),
        "authType":      "",
        "nameRarelyWord": false,
        "mobile":        options.Get("mobile"),
        "sex":           options.Get("passenger_sex"),
        "listNos": []map[string]interface{}{
            {
                "certNo":    options.Get("passenger_idcard"),
                "certType":  1,
                "certName":  "身份证",
                "isDefault": 0,
            },
        },
        "linkerType": 1,
        "linkerName": options.Get("passenger_name"),
        "age":        options.Get("passenger_age"),
    }

	// Marshal the request body into JSON format
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化错误: %v", err)
	}
	jsonstr := string(jsonData)

	// Set request headers
	headers := map[string]string{
		"Cache-Control": "no-cache",
		"Connection": "Keep-Alive",
		"Content-Type": "application/json",
		"Accept": "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		"Host": "wx.17u.cn",
		"Referer": "https://wx.17u.cn/flightbffpassenger/passenger/add",
		"User-Agent": "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Tcplat": "434",
		"Tcuserid": fmt.Sprintf("%s", options.Get("tcuserid")),
		"IsLogin": "true",
		"Diysearch": "1",
		"Tcsectk": fmt.Sprintf("%s", options.Get("secToken")),
		"Tcpolaris": "1",
		"Tcsectoken": fmt.Sprintf("%s", options.Get("tcsectoken")),
		"Tcdeviceid": "433953b3f3805658",
		"Devicetype": "Xiaomi|22041216C",
		"Origin": "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site": "same-origin",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Dest": "empty",
		"Content-Length": "236",
	}

	// Perform the HTTP POST request
	fmt.Println("请求头:", headers)
	fmt.Println("请求体:", jsonstr)

	resp, err := client.
		SetProxyURL("http://127.0.0.1:9000").
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightbffpassenger/passenger/add")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}else {
		fmt.Println("添加乘机人请求成功:", resp.String())
		ListPassenger(options)
	}

	// Return response as string
	return resp.String(), nil
}
func ListPassenger(options Options) (string, error) {
	client := req.C()

	// 构建请求体
	requestBody := map[string]interface{}{
		"source": map[string]interface{}{
			"marketing": map[string]interface{}{
				"bizCode":          "",
				"btpt":            "",
				"mobileLimit":     map[string]interface{}{},
				"studentAuthType": -1,
				"airlineStudent":  false,
				"studentTicket":   false,
				"rels":            []interface{}{},
			},
			"arriveCode":     []string{fmt.Sprintf("%s", options.Get("arrivalCode"))},
			"departCode":     []string{fmt.Sprintf("%s", options.Get("departureCode"))},
			"landDate":      fmt.Sprintf("%s", options.Get("departureDate")),
			"flyDate":       fmt.Sprintf("%s", options.Get("departureDate")),
			"airlineCodeList": []string{"SC", "CA"},
			"isAuthRealname": false,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化错误: %v", err)
	}
	jsonstr := string(jsonData)

	// 设置请求头
	headers := map[string]string{
		"Cache-Control": "no-cache",
		"Connection": "Keep-Alive",
		"Content-Type": "application/json",
		"Accept": "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		"Host": "wx.17u.cn",
		"Referer": "https://wx.17u.cn/flightbffpassenger/passenger/add",
		"User-Agent": "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Tcplat": "434",
		"Tcuserid": fmt.Sprintf("%s", options.Get("tcuserid")),
		"IsLogin": "true",
		"Diysearch": "1",
		"Tcsectk": fmt.Sprintf("%s", options.Get("secToken")),
		"Tcpolaris": "1",
		"Tcsectoken": fmt.Sprintf("%s", options.Get("tcsectoken")),
		"Tcdeviceid": "433953b3f3805658",
		"Devicetype": "Xiaomi|22041216C",
		"Origin": "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site": "same-origin",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Dest": "empty",
		"Content-Length": "236",
	}

	// Perform the HTTP POST request
	fmt.Println("请求头:", headers)
	fmt.Println("请求体:", jsonstr)

	// 发送POST请求
	resp, err := client.
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightbffpassenger/passenger/list")

	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}

	// 检查响应是否成功
	success := gjson.Get(resp.String(), "success").Bool()
	if !success {
		return "", fmt.Errorf("获取乘客列表失败: %s", resp.String())
	}else {
		fmt.Println("获取乘客列表成功:", resp.String())
		GetGSGuid(options)
	}
	
	return resp.String(), nil
}


func GetGSGuid(options Options) (string, error) {
	client := req.C()

	requestBody := map[string]interface{}{
		"dcc": options.Get("departureCode"),
		"acc": options.Get("arrivalCode"),
		"dd": options.Get("departureDate"),
		"pt":  0,
		"dac": options.Get("departureCode"),
		"aac": options.Get("arrivalCode"),
		"fn":  options.Get("flightNo"),
		"clf": 0,
		"entrance": 0,
		"fp": "vb15",
		"passengerInfo": map[string]interface{}{
			"adultCount":  "1",
			"childCount":  "0",
			"infantCount": "0",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化错误: %v", err)
	}
	jsonstr := string(jsonData)

	// 设置请求头
	headers := map[string]string{
		"Host": "wx.17u.cn",
		"Connection": "keep-alive",
		"Content-Length": "236",
		"User-Agent": "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Accept": "application/json, text/plain, */*",
		"Content-Type": "application/json",
		"tcunionid": "",
		"tcplat": "434",
		"tcuserid": fmt.Sprintf("%s", options.Get("tcuserid")),
		"isLogin": "true",
		"diysearch": "1",
		"tcsectk": fmt.Sprintf("%s", options.Get("secToken")),
		"tcpolaris": "1",
		"cache-control": "no-cache",
		"tcsectoken": fmt.Sprintf("%s", options.Get("tcsectoken")),
		"tcdeviceid": "433953b3f3805658",
		"devicetype": "Xiaomi|22041216C",
		"Origin": "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site": "same-origin",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Dest": "empty",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	}

	resp, err := client.
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightbffv2/book15/cabins")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	// 这里可以根据实际响应结构判断是否成功
	success := gjson.Get(resp.String(), "success").Bool()
	if success {
		fmt.Println("获取GSGuid请求成功:", resp.String())
		// fmt.Println("获取GSGuid请求成功")
		// 解析GSGuid
		gsguid := gjson.Get(resp.String(), "data.itinerary.ps.0.tsps.0.tag").String()
		options.Set("GSGuid", gsguid)
		fmt.Println("获取到的GSGuid:", gsguid)
		Buildtemporder(options)
		return resp.String(), nil
	} else {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("获取GSGuid请求失败: %s", resp.String())
	}
}

func Buildtemporder(options Options) (string, error) {
	client := req.C()

	requestBody := map[string]interface{}{
		"FlightType": 1,
		"InsType": 1,
		"IsBackOrder": false,
		"IsMultipleIns": true,
		"IsUnionOrder": false,
		"IsYouXuan": 0,
		"InsurBind": false,
		"IsNeedFlights": true,
		"fromPage": "book15",
		"GSGuid": options.Get("GSGuid"),
		"BSGuid": "",
	  }

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化错误: %v", err)
	}
	jsonstr := string(jsonData)

	// 设置请求头
	headers := map[string]string{
		"Host": "wx.17u.cn",
		"Connection": "keep-alive",
		"Content-Length": "236",
		"User-Agent": "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Accept": "application/json, text/plain, */*",
		"Content-Type": "application/json",
		"tcunionid": "",
		"tcplat": "434",
		"tcuserid": fmt.Sprintf("%s", options.Get("tcuserid")),
		"isLogin": "true",
		"diysearch": "1",
		"tcsectk": fmt.Sprintf("%s", options.Get("secToken")),
		"tcpolaris": "1",
		"cache-control": "no-cache",
		"tcsectoken": fmt.Sprintf("%s", options.Get("tcsectoken")),
		"tcdeviceid": "433953b3f3805658",
		"devicetype": "Xiaomi|22041216C",
		"Origin": "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site": "same-origin",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Dest": "empty",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	}

	resp, err := client.
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightcreateorder/buildtemporder")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	// 这里可以根据实际响应结构判断是否成功
	respc := gjson.Get(resp.String(), "Data.RspCode")
	if respc.String() == "0" {
		fmt.Println("Buildtemporder请求成功:", resp.String())
		fmt.Println("Buildtemporder请求成功")
		// 解析GSGuid
		SerialId := gjson.Get(resp.String(), "Data.SerialId").String()
		backCode := gjson.Get(resp.String(), "Data.Flight.0.cabins.0.cashBackPerOrder.backCode").String()
		options.Set("SerialId", SerialId)
		options.Set("backCode", backCode)
		OrderQuery(options)
		return resp.String(), nil
	} else {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("获Buildtemporder请求失败: %s", resp.String())
	}
}

func OrderQuery(options Options) (string, error) {
	client := req.C()

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	requestBody := map[string]interface{}{
		"channel":    "434",
		"source":     1,
		"unionId":    "",
		"openId":     "",
		"strMemberId": options.Get("tcuserid"),
		"traceId":    "04ecff62-851c-4103-8c32-6f2bbb3f155c",
		"invoker":    "BOOK_2_GIFT",
		"tools":      []string{"mz"},
		"timespan":   timestamp,
		"ext":        map[string]interface{}{},
		"pageId":     "BOOK_2",
		"flight": map[string]interface{}{
			"id":                  "1111",
			"carrier":             []string{"SC"},
			"operateCarrier":      []string{"CA"},
			"departureCityCode":   options.Get("departureCode"),
			"arriveCityCode":       options.Get("arrivalCode"),
			"departureDate":       options.Get("departureDate"),
			"arriveDate":         options.Get("arriveDate"),
			"airlineType":        1,
			"departure":          "SHA",
			"arrival":           "PEK",
			"departureTerminalCode": "T2",
			"arriveTerminalCode":  "T3",
			"products": []map[string]interface{}{
				{
					"id":             "0000",
					"tags":          []string{"00000"},
					"cabin":         []string{fmt.Sprintf("%s", options.Get("cabinCode"))},
					"cabinClass":    []string{fmt.Sprintf("%s", options.Get("cabinClass"))},
					"productPin":    []string{fmt.Sprintf("%s", options.Get("productPin"))},
					"policyType":    41,
					"reimbursement": "2",
					"suppliers":     []string{""},
					"cabinCode":     options.Get("cabinCode"),
					"orderSaleType": 2,
					"productSystem": 2,
					"fare":         options.Get("fare"), //航运价
					"ext": map[string]interface{}{
						"recommendTag": "",
						"taGoods":     "",
					},
					"flightNo": []string{fmt.Sprintf("%s", options.Get("flightNo"))},
				},
			},
			"lcnMapper": map[string]interface{}{
				fmt.Sprintf("%s", options.Get("flightNo")): "A",  //航班号
			},
			"orderAmount": options.Get("orderAmount"),
		},
		"extFlights": nil,
		"passengers": []map[string]interface{}{
			{
				"type":     0,
				"name":     options.Get("passenger_name"),
				"certType": "身份证",
				"certNo":   options.Get("passenger_idcard"),
				"sex":      options.Get("passenger_sex"),
			},
		},
		"trainTicket":     nil,
		"flightTrainFlag": false,
		"specialTags":     []string{"DYNAMIC_PACKAGE"},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化错误: %v", err)
	}
	jsonstr := string(jsonData)
	fmt.Println("确认订单请求体:", jsonstr)
	// 设置请求头
	headers := map[string]string{
		"Host": "wx.17u.cn",
		"Connection": "keep-alive",
		"Content-Length": "236",
		"User-Agent": "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Accept": "application/json, text/plain, */*",
		"Content-Type": "application/json",
		"tcunionid": "",
		"tcplat": "434",
		"tcuserid": fmt.Sprintf("%s", options.Get("tcuserid")),
		"isLogin": "true",
		"diysearch": "1",
		"tcsectk": fmt.Sprintf("%s", options.Get("secToken")),
		"tcpolaris": "1",
		"cache-control": "no-cache",
		"tcsectoken": fmt.Sprintf("%s", options.Get("tcsectoken")),
		"tcdeviceid": "433953b3f3805658",
		"devicetype": "Xiaomi|22041216C",
		"Origin": "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site": "same-origin",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Dest": "empty",
		"Accept-Language": "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	}

	resp, err := client.
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightedward/whosyourdaddy/promotion/query")

	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}

	// 这里可以根据实际响应结构判断是否成功
	success := gjson.Get(resp.String(), "data.success").Bool()
	if success {
		fmt.Println("订单查询请求成功:", resp.String())
		traceId := gjson.Get(resp.String(), "data.traceId").String()
		code := gjson.Get(resp.String(), "data.promotions.0.code").String()
		promotionSign := gjson.Get(resp.String(), "data.promotions.0.promotionSign").String()
		options.Set("traceId", traceId)
		options.Set("code", code)
		options.Set("promotionSign", promotionSign)
		CreateOrder(options)

		return resp.String(), nil
	} else {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("订单查询请求失败: %s", resp.String())
	}
}

func CreateOrder(options Options) (string, error) {
	client := req.C()

	// 构建请求体
	requestBody := map[string]interface{}{
		"OrderSerialId": options.Get("OrderSerialId"),
		"isNewInsuranceType": 1,
		"LinkMobile": fmt.Sprintf("%s", options.Get("mobile")),
		"opsArr": []map[string]interface{}{
			{
				"type":       1,
				"name":       fmt.Sprintf("%s", options.Get("passenger_name")),
				"certName":   "身份证",
				"certNo":    fmt.Sprintf("%s", options.Get("passenger_no")),
				"birthDay":  fmt.Sprintf("%s", options.Get("birthday")),
				"gender":     1,
				"passId":    fmt.Sprintf("%s", options.Get("passenger_idcard")),
				"linkPhone": fmt.Sprintf("%s", options.Get("mobile")),
				"attr":      []interface{}{},
				"country":    "",
				"surname":    "",
				"givenName":  "",
				"dateOfExpiry": "",
				"memberType": "0",
			},
		},
		"IsNeedSend": "0",
		"IsRegTcMember": true,
		"IsCheckCertNo": 1,
		"EnsurePassageInfoStr": `{"isEnsure":0,"EnsurePassage":[]}`,
		"UCType": "",
		"ErrorType": 1,
		"SegmentType": 1,
		"gwPassengerLimitSwitch": false,
		"flightTicketVoucher": false,
		"reimbursementType": 0,
		"marks": []string{"NEW_INSURANCE_FLOW"},
		"packOrderUnionNo": "",
		"isSpecialMember": 0,
		"ancillaryRights": []interface{}{},
		"outRefid": "",
		"BackOrderSerialId": nil,
		"LinkMan": "许玉鹏",
		"LinkCertNo": "",
		"activities": []map[string]interface{}{
			{
				"code":   "28739",
				"type":   "PRE_ORDER_CASH_BACK",
				"amount": 7,
			},
		},
		"unifiedMileages": []interface{}{},
		"GiftCodes": []map[string]interface{}{
			{
				"type":         0,
				"code":        options.Get("code"),
				"price":       0,
				"promotionSign": options.Get("promotionSign"),
			},
		},
		"ClientToken": `{"MailABKey":0,"HasMailCheck":1,"cabinsProductType":0,"cabinTypeName":"同程特惠","OrderCash":"A"}`,
		"commonProductList": []interface{}{},
		"orderTotalPrice": 848,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化错误: %v", err)
	}
	jsonstr := string(jsonData)

	// 设置请求头
	headers := map[string]string{
		"Host":             "wx.17u.cn",
		"Connection":       "keep-alive",
		"Content-Length":   fmt.Sprintf("%d", len(jsonstr)),
		"User-Agent":       "Mozilla/5.0 (Linux; Android 13; 22041216C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/104.0.5112.97 Mobile Safari/537.36/TcTravel/11.0.6",
		"Accept":           "application/json, text/plain, */*",
		"Content-Type":     "application/json",
		"tcunionid":        "",
		"tcplat":           "434",
		"tcuserid":         fmt.Sprintf("%s", options.Get("tcuserid")),
		"isLogin":          "true",
		"diysearch":        "1",
		"tcsectk":          fmt.Sprintf("%s", options.Get("sec_token")),
		"tcpolaris":        "1",
		"cache-control":    "no-cache",
		"tcsectoken":       fmt.Sprintf("%s", options.Get("tcsectoken")),
		"tcdeviceid":       "433953b3f3805658",
		"devicetype":       "Xiaomi|22041216C",
		"Origin":           "https://wx.17u.cn",
		"X-Requested-With": "com.tongcheng.android",
		"Sec-Fetch-Site":   "same-origin",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
		"Accept-Language":  "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
	}

	// 发送POST请求
	resp, err := client.
		R().
		SetHeaders(headers).
		SetBody(jsonstr).
		Post("https://wx.17u.cn/flightbfforder/create/order/unity")

	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}

	// 检查响应是否成功
	success := gjson.Get(resp.String(), "success").Bool()
	if !success {
		fmt.Println("请求失败:", resp.String())
		return "", fmt.Errorf("创建订单失败: %s", resp.String())
	}
	fmt.Println("创建订单请求成功:", resp.String())
	return resp.String(), nil
}

