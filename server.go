package main

import (
	"container/list"
	"flag"
	_ "fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
	qrcode "github.com/skip2/go-qrcode"
	_ "html/template"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	_ "os"
	"strconv"
	"strings"
	"time"
)

func qr(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	param, _ := req.Form["data"]
	if len(param) != 1 {
		return
	}
	temp := strings.Replace(param[0], "!", "&", -1)
	var png []byte
	png, err := qrcode.Encode(string(temp), qrcode.Medium, 256)

	if err != nil {
		log.Println("qrcode error:", err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)

}

func KEY() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//log.Println(KEY())

func timestamp() float64 {
	var timestamp = float64(time.Now().UnixNano())
	return timestamp / 1e6
}

func home(w http.ResponseWriter, r *http.Request) {
	t, _ := ioutil.ReadFile("./templates/intro.html")
	w.Write([]byte(t))
}
func pannel(w http.ResponseWriter, r *http.Request) {
	t, _ := ioutil.ReadFile("./templates/pannel.html")
	w.Write([]byte(t))
}

type verify_roomtoken_struct struct {
	success     int
	description string
}

func verify_roomtoken(roomid int, roomtoken string) verify_roomtoken_struct {
	var t verify_roomtoken_struct
	temp, err := roomdic.HGet(strconv.Itoa(roomid),"roomtoken").Result() 
	if err == redis.Nil { 
		t.success = 0 
		t.description="room not found"
		return t
	}
	if temp!=roomtoken{
		t.success = 0 
		t.description="roomtoken not match"
		log.Println(temp," ",roomtoken)
		return t
	}
	t.success = 1
	return t
}
func canvas(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	roomid := v.Get("roomid")
	roomtoken := v.Get("token")
	if roomid == "" {
		w.Write([]byte("缺少参数"))
		return
	}
	if roomtoken == "" {
		w.Write([]byte("缺少参数"))
		return
	}
	
	T, _ := strconv.Atoi(roomid)
	temp := verify_roomtoken(T, roomtoken)
	log.Println(temp)
	if temp.success==0 {
		w.Write([]byte(temp.description))
		return
	}
	t, _ := ioutil.ReadFile("./templates/canvas.html")
	w.Write([]byte(t))

}

func api(w http.ResponseWriter, r *http.Request) {
	action := mux.Vars(r)["action"]
	v := r.URL.Query()
	//token := v.Get("token")
	roomid, _ := strconv.Atoi(v.Get("roomid"))
	log.Println("api gets called: ", action)
	switch action {
	case "settextcolor":
		color := strings.Replace(v.Get("color"), "@", "#", -1)
		roomdic.HSet(strconv.Itoa(roomid), "textcolor", color)
		broadcast_settextcolor(roomid)
		broadcast_updatepannelui(roomid)
		break

	case "setbackgroundcolor":
		color := strings.Replace(v.Get("color"), "@", "#", -1)
		roomdic.HSet(strconv.Itoa(roomid), "backgroundcolor", color)
		broadcast_setbackgroundcolor(roomid)
		broadcast_updatepannelui(roomid)
		break

	case "settextsize":
		size := v.Get("size")
		roomdic.HSet(strconv.Itoa(roomid), "textsize", size)
		broadcast_settext(roomid)
		broadcast_updatepannelui(roomid)
		break

	case "settext":
		text := v.Get("text")
		roomdic.HSet(strconv.Itoa(roomid), "text", text)
		broadcast_settext(roomid)
		broadcast_updatepannelui(roomid)
		break

	case "setspeed":
		speed := v.Get("speed")
		roomdic.HSet(strconv.Itoa(roomid), "speed", speed)
		broadcast_setspeed(roomid)
		broadcast_updatepannelui(roomid)
		break
	case "setbgblinkinterval":
		interval := v.Get("interval")
		roomdic.HSet(strconv.Itoa(roomid), "bgblinkinterval", interval)
		broadcast_setbgblinkinterval(roomid)
		broadcast_updatepannelui(roomid)
		break
	case "seteffect":
		effect := v.Get("effect")
		roomdic.HSet(strconv.Itoa(roomid), "effect", effect)
		broadcast_seteffect(roomid)
		broadcast_updatepannelui(roomid)
		break
	default:
		w.Write([]byte("invaild method"))
	}

}

func getconfigapi(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	roomid, _ := strconv.Atoi(v.Get("roomid"))
	temp, _ := roomdic.HGetAll(strconv.Itoa(roomid)).Result()
	data, _ := json.Marshal(&temp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type setidstruct struct {
	Action string `json:"action"`
	Id     int    `json:"id"`
	Token  string `json:"token"`
}

func setid(ws wspatch_struct,token string) {
	var m setidstruct
	m.Action = "setid"
	m.Id = ws.no
	m.Token = token
	data, _ := json.Marshal(m)
	send(ws,data)
}

type setstarttimestampstruct struct {
	Action string `json:"action"`
	Timestamp     float64    `json:"timestamp"`
}

func setstarttimestamp(ws wspatch_struct, timestamp float64) {
	var m setstarttimestampstruct
	m.Action = "setstarttimestamp"
	m.Timestamp = timestamp
	data, _ := json.Marshal(m)
	send(ws,data)
}
type setspeedstruct struct {
	Action string  `json:"action"`
	Speed  float64 `json:"speed"`
}

func setspeed(ws wspatch_struct) {
	var m setspeedstruct
	m.Action = "setspeed"
	speedtemp, _ := roomdic.HGet(strconv.Itoa(ws.roomid), "speed").Result()
	speed, _ := strconv.ParseFloat(speedtemp, 64)
	m.Speed = math.Abs(speed)
	data, _ := json.Marshal(m)
	send(ws,data)
}
type setbgblinkintervalstruct struct{
	Action string `json:"action"`
	Interval int `json:"interval"`
}
func setbgblinkinterval(ws wspatch_struct) {
	var m setbgblinkintervalstruct
	m.Action = "setbgblinkinterval"
	intervaltemp, _ := roomdic.HGet(strconv.Itoa(ws.roomid), "bgblinkinterval").Result()
	interval, _ := strconv.Atoi(intervaltemp)
	m.Interval =interval
	data, _ := json.Marshal(m)
	send(ws,data)
}
type settextcolorstruct struct {
	Action string `json:"action"`
	Color  string `json:"color"`
}

func settextcolor(ws wspatch_struct) {
	var m settextcolorstruct
	m.Action = "settextcolor"
	color, _ := roomdic.HGet(strconv.Itoa(ws.roomid), "textcolor").Result()
	m.Color = color
	data, _ := json.Marshal(m)
	send(ws,data)
}

type setbackgroundcolorstruct struct {
	Action string `json:"action"`
	Color  string `json:"color"`
}

func setbackgroundcolor(ws wspatch_struct) {
	var m setbackgroundcolorstruct
	m.Action = "setbackgroundcolor"
	color, _ := roomdic.HGet(strconv.Itoa(ws.roomid), "backgroundcolor").Result()
	m.Color = color
	data, _ := json.Marshal(m)
	send(ws,data)
}

type settextstruct struct {
	Action string  `json:"action"`
	Text   string  `json:"text"`
	Size   float64 `json:"size"`
}

func settext(ws wspatch_struct) {
	var m settextstruct
	m.Action = "settext"
	text, _ := roomdic.HGet(strconv.Itoa(ws.roomid), "text").Result()
	m.Text = text
	sizetemp, _ := roomdic.HGet(strconv.Itoa(ws.roomid), "textsize").Result()
	size, _ := strconv.ParseFloat(sizetemp, 64)
	m.Size = size
	data, _ := json.Marshal(m)
	send(ws, data)
}

type setposstruct struct {
	Action    string  `json:"action"`
	Towards   string  `json:"towards"`
	Px        float64 `json:"px"`
	Timestamp float64 `json:"timestamp"`
}

func setpos(ws wspatch_struct, towards string, px float64, timestamp float64) {
	var m setposstruct
	m.Action = "setpos"
	m.Towards = towards
	m.Px = px
	m.Timestamp = timestamp
	data, _ := json.Marshal(m)
	send(ws, data)
}

type sendmsgstruct struct {
	Action string `json:"action"`
	Msg    string `json:"msg"`
}

func sendmsg(ws wspatch_struct, msg string) {
	var m sendmsgstruct
	m.Action = "showmsg"
	m.Msg = msg
	data, _ := json.Marshal(m)
	send(ws,data)
}
type forcequitstruct struct {
	Action string `json:"action"`
}
func forcequit(ws wspatch_struct) {
	var m singleforceresyncstruct
	m.Action = "forcequit"
	data, _ := json.Marshal(m)
	send(ws,data)
}
type singleforceresyncstruct struct {
	Action string `json:"action"`
}
func forceresync(ws wspatch_struct) {
	var m singleforceresyncstruct
	m.Action = "forceresync"
	data, _ := json.Marshal(m)
	send(ws,data)
}

type singleresyncstruct struct {
	Action string `json:"action"`
}

func resync(ws wspatch_struct) {
	var m singleforceresyncstruct
	m.Action = "resync"
	data, _ := json.Marshal(m)
	send(ws, data)
}

type sendpingstruct struct {
	Action string `json:"action"`
}

func sendping(ws wspatch_struct) {
	var m sendpingstruct
	m.Action = "ping"
	data, _ := json.Marshal(m)
	send(ws, data)
}

func setselfindex(ws wspatch_struct) {
	var m selfindexdatastruct
	m.Action = "setindex"
	m.Index = getselfindex(ws.roomid, ws.no) + 1
	data, _ := json.Marshal(m)
	send(ws,data)
}
type settimestampoffsetstruct struct {
	Action string `json:"action"`
	Offset float64 `json:"offset"`
}

func settimestampoffset(ws wspatch_struct, offset float64) {
	var m settimestampoffsetstruct
	m.Action = "settimestampoffset"
	m.Offset = offset
	data, _ := json.Marshal(m)
	send(ws, data)
}
func send(c wspatch_struct, data []byte) {
	defer func() {
        //恢复程序的控制权
        err := recover()
        if err != nil {
            log.Println("recovered from concurrent write")
        }
    }()
	
	if c.Lock {
		log.Println("send blocked by lock")
	//	for {
	//		time.Sleep(100 * time.Millisecond)
	//		send(c, no, data)
	//	}
		return
	}
	log.Println("sending:", string(data))
	c.Lock = true
	err := c.WS.WriteMessage(websocket.TextMessage, data)
	c.Lock = false
	//[]byte("1")
	if err != nil {
		
		log.Println("thread send error:", err, "\n userid:", c.no, " roomid:", c.roomid)

		userdic.Publish(strconv.Itoa(c.no), "OFFLINE")
		userdic.Del(strconv.Itoa(c.no))
		userlist.LRem(strconv.Itoa(c.roomid), 0, c.no)
		devicedic.Del(strconv.Itoa(c.no))
		broadcast_cycledata(c.roomid)
		broadcast_selfindex(c.roomid)
		broadcast_forceresync(c.roomid)
		broadcast_sendmsg(c.roomid, "一个用户下线")

		if isroomempty(c.roomid) {
			roomdic.Del(strconv.Itoa(c.roomid))
			log.Println("房间已被删除")
		}

	}

}

type wspatch_struct struct {
	Lock bool
	WS   *websocket.Conn
	no int
	roomid int
}

func echo(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws_temp, err := upgrader.Upgrade(w, r, nil)

	var ws wspatch_struct
	ws.Lock = false
	ws.WS = ws_temp

	if err != nil {
		log.Print("upgrade to websocket failed", err)
		return
	}
	defer ws.WS.Close()
	var (
		no                 int
		roomid             int
		latency            float64
		online             bool    = true
		lastping_timestamp float64 = timestamp()
	)

	for {
		_, data, err := ws.WS.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			log.Println("USER OFFLINE")
			break
		}
		log.Printf("recv: %s", data)

		//////////////////////////////////////////////
		go func() {

			var parsed map[string]interface{}
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				log.Println("parse websocket json failed", err)
				return
			}

			switch parsed["action"] {

			case "statusreport":
			//	if parsed["isrendering"].(bool) {
					p := selfpos(roomid, no, -latency)
					T := p.Front().Value.(float64)
					if parsed["current"]==nil {
						break
					}
					if math.Abs(parsed["current"].(float64)-T) > 5 {
						log.Println("误差过大", parsed["current"].(float64)-T)
						resync(ws)
					}
					break
				//}

			case "reporttextwidth":
				devicedic.HSet(strconv.Itoa(no), "textwidth", parsed["textwidth"])
				broadcast_cycledata(roomid)
				break

			case "reportscreenwidth":
				devicedic.HSet(strconv.Itoa(no), "screenwidth", parsed["screenwidth"])
				broadcast_cycledata(roomid)
				broadcast_forceresync(roomid)
				break

			case "resync":
				t := selfpos(roomid, no, latency)
				px := t.Front().Value.(float64)
				towards := t.Back().Value.(string)
				setpos(ws, towards, px, timestamp())
				break

			case "pong":
				latency = (timestamp() - lastping_timestamp) / 2
				log.Println("延迟:", latency)
				if latency > 300 {
					sendmsg(ws, "延迟过高")
				}
				error_in_ms:=timestamp()-(parsed["timestamp"].(float64)+latency)
        	    if math.Abs(error_in_ms)>5{
        	  	settimestampoffset(ws,error_in_ms)
                }
			case "init":
				roomid_temp := parsed["roomid"].(string)
				roomtoken := parsed["roomtoken"].(string)
				roomid, _ = strconv.Atoi(roomid_temp)
				ws.roomid=roomid
				verify_roomtoken_result:=verify_roomtoken(roomid,roomtoken)
				if verify_roomtoken_result.success==0{
					forcequit(ws)
					return
				}
				
				userdic.Incr("usercount")
				no_temp, _ := userdic.Get("usercount").Result()
				no, _ = strconv.Atoi(no_temp)
				ws.no=no
				token := KEY()
				setid(ws, token)
				/////
				start_timestamp_temp, _ := roomdic.HGet(strconv.Itoa(roomid), "start_timestamp").Result()
	            start_timestamp, _ := strconv.ParseFloat(start_timestamp_temp, 64)
				setstarttimestamp(ws, start_timestamp)
				////
				setspeed(ws)
				setbgblinkinterval(ws)
				settextcolor(ws)
				setbackgroundcolor(ws)
				settext(ws)

				go func() {
					time.Sleep(500 * time.Millisecond)
					forceresync(ws)
				}()

				go func() {

					for {
						pubsub := userdic.Subscribe("room_"+strconv.Itoa(roomid), strconv.Itoa(no))
						//defer pubsub.Close()
						_, err := pubsub.Receive()
						if err != nil {
							return
						}
						ch := pubsub.Channel()
						for msg := range ch {
							log.Println("broadcast received：", msg.Channel, msg.Payload)
							if msg.Payload == "OFFLINE" {
								log.Println("广播监听线程退出")
								online = false
								pubsub.Close()
								return
							}
							if msg.Channel == "room_"+strconv.Itoa(roomid) {
								var t map[string]interface{}
								json.Unmarshal([]byte(msg.Payload), &t)
								if t["action"] == "totarget" && (t["target"] == float64(no) || t["target"] == "all") {
									log.Println("BROADCASTING:", t["data"])
									send(ws, []byte(t["data"].(string)))
								}

							}

						}
					}
				}()
				go func() {
					for {
						time.Sleep(500 * time.Millisecond)
						if !online {
							return
						}
						sendping(ws)
						lastping_timestamp = timestamp()
					}
				}()
				//////////	DEBUG ONLY
				//	go func(){
				//	for{
				//	    time.Sleep(16*time.Millisecond)
				//		T:=get_pos_incavas(roomid,0)
				//		log.Println(T.Front().Value.(float64),T.Back().Value.(string),"!!!!!!!!!")
				//		}
				//	}()
				/////////////
				devicedic.HSet(strconv.Itoa(no), "ua", parsed["ua"])
				devicedic.HSet(strconv.Itoa(no), "token", token)
				devicedic.HSet(strconv.Itoa(no), "screenwidth", parsed["screenwidth"])
				userlist.RPush(strconv.Itoa(roomid), no)
				userdic.Set(strconv.Itoa(no), roomid, -1)
				setselfindex(ws)
				broadcast_forceresync(roomid)
				broadcast_sendmsg(roomid, "一个新用户成功加入")

			default:
				panic("JSON not understood")
			}
		}()
		//////////////////////////////////////

	}
}

func index(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}
func getuserlist(roomid int) []string {
	result, _ := userlist.LRange(strconv.Itoa(roomid), 0, -1).Result()
	return result
}
func isroomempty(roomid int) bool {
	t := getuserlist(roomid)
	return len(t) == 0
}
func ismaster(roomid int, no int) bool {
	return getselfindex(roomid, no) == 0
}
func getselfindex(roomid int, no int) int {
	var temp []string = getuserlist(roomid)
	return index(temp, strconv.Itoa(no))
}
func getuserbefore(roomid int, no int, onlyone bool) []string {
	var temp []string = getuserlist(roomid)
	var t int = index(temp, strconv.Itoa(no)) - 1
	if t < 0 {
		temp = []string{}
		return temp
	}
	if onlyone {
		temp = []string{temp[t]}
		return temp
	}
	return temp[:t+1]
}

func getuserafter(roomid int, no int, onlyone bool) []string {
	var temp []string = getuserlist(roomid)
	var t int = index(temp, strconv.Itoa(no)) + 1
	if t < 0 {
		temp = []string{}
		return temp
	}
	if onlyone {
		temp = []string{temp[t]}
		return temp
	}
	return temp[t:]

}

func gettextwidth(no int) float64 {
	result, err := devicedic.HGet(strconv.Itoa(no), "textwidth").Result()
	if err != nil {
		log.Println("TEXTWIDTH READ FAILURE")
		return 0
	}
	T, _ := strconv.ParseFloat(result, 64)
	return T
}
func getscreenwidth(no int) float64 {
	result, err := devicedic.HGet(strconv.Itoa(no), "screenwidth").Result()
	if err != nil {
		log.Println("TEXTWIDTH READ FAILURE")
		return 0
	}
	T, _ := strconv.ParseFloat(result, 64)
	return T
}
func getwidth(t []string) *list.List {
	l := list.New()
	for i := 0; i < len(t); i++ {
		no, _ := strconv.Atoi(t[i])
		l.PushBack(getscreenwidth(no))
	}
	return l
}
func getdevicewidthlist(roomid int) *list.List {
	return getwidth(getuserlist(roomid))
}
func sum(l *list.List) float64 {
	var sum float64 = 0
	for item := l.Front(); nil != item; item = item.Next() {
		sum += item.Value.(float64)
	}
	return sum
}
func getcanvaswidth(roomid int) float64 {
	return sum(getdevicewidthlist(roomid))
}
func getwidthlistafteruser(roomid int, no int) *list.List {
	return getwidth(getuserafter(roomid, no, false))
}
func getwidthlistbeforeuser(roomid int, no int) *list.List {
	return getwidth(getuserbefore(roomid, no, false))
}
func getsafezone(roomid int) float64 {
	t:=getuserlist(roomid)
	if len(t)==0{return 0}
	temp, _ := strconv.Atoi(getuserlist(roomid)[0])
	return gettextwidth(temp)
}
func getcycledata(roomid int, no int) [2]float64 {
	log.Println(getwidthlistbeforeuser(roomid, no), getwidthlistafteruser(roomid, no), no)
	rightwidth := sum(getwidthlistafteruser(roomid, no))
	leftwidth := sum(getwidthlistbeforeuser(roomid, no)) + getsafezone(roomid)
	log.Println(leftwidth, rightwidth)
	result := [2]float64{leftwidth, rightwidth}
	return result
}
func get_pos_incavas(roomid int, latency float64) *list.List {

	now := timestamp()
	speedtemp, _ := roomdic.HGet(strconv.Itoa(roomid), "speed").Result()
	speed, _ := strconv.ParseFloat(speedtemp, 64)
	start_timestamp_temp, _ := roomdic.HGet(strconv.Itoa(roomid), "start_timestamp").Result()
	start_timestamp, _ := strconv.ParseFloat(start_timestamp_temp, 64)
	ms_passed := now - start_timestamp + latency
	movement_per_ms := math.Abs(speed) * 60 / 1000
	total_movement := ms_passed * movement_per_ms
	safezone := getsafezone(roomid)
	width := getcanvaswidth(roomid) + safezone
	temp := int(total_movement / width)
	effect, _ := roomdic.HGet(strconv.Itoa(roomid), "effect").Result()
	var (
		valid_movement float64
		towards        float64
		result         float64
		towardstr      string
	)
	switch effect {
	case "normal":
		valid_movement = float64(total_movement - width*float64(temp))
		if speed < 0 {
			valid_movement = width - valid_movement
		}
		towards = speed
		break
	case "bounce":
		valid_movement = float64(total_movement - width*float64(temp))
		if temp%2 == 0 {
			if speed < 0 {
				valid_movement = width - valid_movement
			}
			towards = speed
		} else {
			if speed > 0 {
				valid_movement = width - valid_movement
			}
			towards = -speed
		}
		break
	}

	if valid_movement >= 0 {
		result = valid_movement - safezone
	} else {
		result = width + valid_movement - safezone
	}
	if towards > 0 {
		towardstr = "right"
	} else {
		towardstr = "left"
	}
	l := list.New()
	l.PushBack(result)
	l.PushBack(towardstr)
	// l.PushBack(temp)
	return l
}
func selfpos(roomid int, no int, latency float64) *list.List {
	px_before := sum(getwidthlistbeforeuser(roomid, no))
	//selfwidth:=getscreenwidth(no)
	pos := get_pos_incavas(roomid, latency)
	temp := pos.Front().Value.(float64) - px_before

	l := list.New()
	l.PushBack(temp)
	l.PushBack(pos.Back().Value.(string))
	return l
}

//////////////////////////////////////////////////
type seteffectstruct struct {
	Action string `json:"action"`
	Effect string `json:"effect"`
}

func broadcast_seteffect(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 seteffectstruct
	m2.Action = "seteffect"
	effect, _ := roomdic.HGet(strconv.Itoa(roomid), "effect").Result()
	m2.Effect = effect
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

//////////////////////////////////////////////////
func broadcast_setspeed(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 setspeedstruct
	m2.Action = "setspeed"
	speedtemp, _ := roomdic.HGet(strconv.Itoa(roomid), "speed").Result()
	speed, _ := strconv.ParseFloat(speedtemp, 64)
	m2.Speed = math.Abs(speed)
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

//////////////////////////////////////////////////
func broadcast_setbgblinkinterval(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 setbgblinkintervalstruct
	m2.Action = "setbgblinkinterval"
	intervaltemp, _ := roomdic.HGet(strconv.Itoa(roomid), "bgblinkinterval").Result()
	interval, _ := strconv.Atoi(intervaltemp)
	m2.Interval = interval
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}
//////////////////////////////////////////////////
func broadcast_settext(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 settextstruct
	m2.Action = "settext"
	text, _ := roomdic.HGet(strconv.Itoa(roomid), "text").Result()
	m2.Text = text
	sizetemp, _ := roomdic.HGet(strconv.Itoa(roomid), "textsize").Result()
	size, _ := strconv.ParseFloat(sizetemp, 64)
	m2.Size = size
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

//////////////////////////////////////////////////
//////////////////////////////////////////////////
type updatepanneluistruct struct {
	Action string `json:"action"`
}

func broadcast_updatepannelui(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 updatepanneluistruct
	m2.Action = "updatepannelui"
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

//////////////////////////////////////////////////
func broadcast_setbackgroundcolor(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 setbackgroundcolorstruct
	m2.Action = "setbackgroundcolor"
	color, _ := roomdic.HGet(strconv.Itoa(roomid), "backgroundcolor").Result()
	m2.Color = color
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

//////////////////////////////////////////////////
func broadcast_settextcolor(roomid int) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 settextcolorstruct
	m2.Action = "settextcolor"
	color, _ := roomdic.HGet(strconv.Itoa(roomid), "textcolor").Result()
	m2.Color = color
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

//////////////////////////////////////////////////
type cycledatastruct struct {
	Action string     `json:"action"`
	Data   [2]float64 `json:"data"`
}
type broadcaststruct struct {
	Action string `json:"action"`
	Target int    `json:"target"`
	Data   string `json:"data"`
}
type broadcaststruct2 struct {
	Action string `json:"action"`
	Target string `json:"target"`
	Data   string `json:"data"`
}

func broadcast_cycledata(roomid int) {
	t := getuserlist(roomid)
	for i := 0; i < len(t); i++ {
		no, _ := strconv.Atoi(t[i])
		var m1 broadcaststruct
		m1.Action = "totarget"
		m1.Target = no
		var m2 cycledatastruct
		m2.Action = "setcycledata"
		m2.Data = getcycledata(roomid, no)
		data, _ := json.Marshal(m2)
		m1.Data = string(data)
		data, _ = json.Marshal(m1)
		err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
		if err != nil {
			log.Println("发布失败")
			return
		}
	}
}

/////////////////////////////////////////////////////
//////////////////////////////////////////////////
type selfindexdatastruct struct {
	Action string `json:"action"`
	Index  int    `json:"index"`
}

func broadcast_selfindex(roomid int) {
	t := getuserlist(roomid)
	for i := 0; i < len(t); i++ {
		no, _ := strconv.Atoi(t[i])
		var m1 broadcaststruct
		m1.Action = "totarget"
		m1.Target = no
		var m2 selfindexdatastruct
		m2.Action = "setindex"
		m2.Index = getselfindex(roomid, no) + 1
		data, _ := json.Marshal(m2)
		m1.Data = string(data)
		data, _ = json.Marshal(m1)
		err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
		if err != nil {
			log.Println("发布失败")
			return
		}
	}
}

/////////////////////////////////////////////////////
//////////////////////////////////////////////////
type forceresyncstruct struct {
	Action string `json:"action"`
}

func broadcast_forceresync(roomid int) {

	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 forceresyncstruct
	m2.Action = "forceresync"
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

/////////////////////////////////////////////////////
/////////////////////////////////////////////////////

func broadcast_sendmsg(roomid int, msg string) {
	var m1 broadcaststruct2
	m1.Action = "totarget"
	m1.Target = "all"
	var m2 sendmsgstruct
	m2.Action = "showmsg"
	m2.Msg = msg
	data, _ := json.Marshal(m2)
	m1.Data = string(data)
	data, _ = json.Marshal(m1)
	err := userdic.Publish("room_"+strconv.Itoa(roomid), data).Err()
	if err != nil {
		log.Println("发布失败")
		return
	}
}

////////////////////////////////////////////////////
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var addr = flag.String("addr", "localhost:6000", "http service address")
var upgrader = websocket.Upgrader{}

var redisAddr = "localhost:6379"
var redisPassword = "wfkycEzzxjl1"
var userdic = redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPassword, DB: 10})
var devicedic = redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPassword, DB: 11})
var userlist = redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPassword, DB: 12})
var roomdic = redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPassword, DB: 13})

func main() {
    log.Print("mkdm serving at port 6000")
	userdic.FlushDB()
	devicedic.FlushDB()
	userlist.FlushDB()
	roomdic.FlushDB()
	
	/*
	   roomdic.HSet(strconv.Itoa(0),"text","abcd")
	   roomdic.HSet(strconv.Itoa(0),"speed",-10)
	   roomdic.HSet(strconv.Itoa(0),"start_timestamp",timestamp())
	   roomdic.HSet(strconv.Itoa(0),"musicbegin_timestamp",timestamp())
	   roomdic.HSet(strconv.Itoa(0),"textsize",400)
	   roomdic.HSet(strconv.Itoa(0),"textcolor","#006400")
	   roomdic.HSet(strconv.Itoa(0),"backgroundcolor","#FFFFFF")
	   roomdic.HSet(strconv.Itoa(0),"effect","bounce")
	   roomdic.HSet(strconv.Itoa(0),"roomtoken",KEY())
	*/

	flag.Parse()
	log.SetFlags(0)

	mux := mux.NewRouter()
	mux.HandleFunc("/ws", echo)
	mux.HandleFunc("/", home)
	mux.HandleFunc("/qr", qr)
	mux.HandleFunc("/api/{action}", api)
	mux.HandleFunc("/createroom", createroom)
	mux.HandleFunc("/getconfig", getconfigapi)
	mux.HandleFunc("/pannel", pannel)
	mux.HandleFunc("/canvas", canvas)

	log.Fatal(http.ListenAndServe(*addr, mux))
}

type createroom_struct struct {
	Roomid int    `json:"roomid"`
	Token  string `json:"token"`
}

func createroom(w http.ResponseWriter, r *http.Request) {

	token := KEY()
	userdic.Incr("roomcount")
	roomid_str, _ := userdic.Get("roomcount").Result()
	roomid, _ := strconv.Atoi(roomid_str)

	roomdic.HSet(roomid_str, "text", "Milky Way Barrage")
	roomdic.HSet(roomid_str, "speed", -10)
	roomdic.HSet(roomid_str, "start_timestamp", timestamp())
	roomdic.HSet(roomid_str, "musicbegin_timestamp", timestamp())
	roomdic.HSet(roomid_str, "textsize", 400)
	roomdic.HSet(roomid_str, "textcolor", "#006400")
	roomdic.HSet(roomid_str, "backgroundcolor", "#FFFFFF")
	roomdic.HSet(roomid_str, "effect", "bounce")
	roomdic.HSet(roomid_str, "roomtoken", token)
	roomdic.HSet(roomid_str, "bgblinkinterval", 0)
	

	var m createroom_struct
	m.Roomid = roomid
	m.Token = token
	data, _ := json.Marshal(&m)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
