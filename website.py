from flask import Flask, render_template, url_for,request,jsonify,make_response,abort,redirect,send_file
import json,time,multiprocessing,threading,random,string
import qrcode
from io import BytesIO

app = Flask(__name__ ,static_folder='templates',static_url_path='')

#########################################################################################
def KEY():
    keylist = [random.choice(string.ascii_letters+string.digits) for i in range(32)]
    return ("".join(keylist))



######################################### 以下为redis内存数据库############################

import redis
redisdic={10:"userdic",11:"devicedic",12:"userlist",13:"roomdic"}

redis.ConnectionPool(host='localhost', port=6379, password='wfkycEzzxjl1',decode_responses=True)

for key,value in redisdic.items():
   exec(value+"""= redis.Redis(connection_pool=redis.ConnectionPool(host='localhost', port=6379, password='wfkycEzzxjl1',decode_responses=True,db="""+str(key)+"""))
"""+value+""".flushdb() 
"""+value+""".hmset("startuptest", {"test": "redis内存数据库启动！msg from:  """+value+""""})
print("""+value+""".hget("startuptest", "test"))
""")
print("内存数据库启动完成")

#########################################
def verify_roomtoken(roomid,roomtoken):
	if not roomid or not roomtoken:
		return {"success":0,"description":"缺少参数"}
	if not roomdic.exists(roomid):
		return {"success":0,"description":"room not found"}
	if not roomdic.hget(roomid,"roomtoken")==roomtoken:
		return {"success":0,"description":"roomtoken not match"}
	return	{"success":1}


@app.route("/qr")
def qr():
    data=request.args.get("data").replace("!","&")
    img_buf = BytesIO()
    qr = qrcode.QRCode(version=1,
                       error_correction=qrcode.constants.ERROR_CORRECT_L,
                       box_size=10,
                       border=4)

    qr.add_data(data)
    qr.make(fit=True)
    img = qr.make_image()
    img.save(img_buf)
    img_buf.seek(0)
    return send_file(img_buf, mimetype='image/png')	
    
@app.route("/")
def main():
	return  render_template("/intro.html")
@app.route("/old")
def mainold():
    return  render_template("/canvasold.html")
@app.route("/pannel")
def pannel():
    return  render_template("/pannel.html")
@app.route("/canvas")
def canvas():
	roomid=request.args.get("roomid")
	roomtoken=request.args.get("token")
	temp=verify_roomtoken(roomid,roomtoken)
	if not temp["success"]:
		return temp["description"]
	return  render_template("/canvas.html")    
@app.route('/getconfig')
def getconfigapi():
    roomid=request.args.get("roomid")
    return jsonify(roomdic.hgetall(roomid))


@app.route("/createroom")
def createroom():
	token=KEY()
	userdic.incr("roomcount")
	roomid=userdic.get("roomcount")
	
	roomdic.hset(roomid,"text","Milky Way Barrage")
	roomdic.hset(roomid,"speed",-10)
	roomdic.hset(roomid,"start_timestamp",time.time() * 1000)
	roomdic.hset(roomid,"musicbegin_timestamp",time.time() * 1000)
	roomdic.hset(roomid,"textsize",400)
	roomdic.hset(roomid,"textcolor","#006400")
	roomdic.hset(roomid,"backgroundcolor","#FFFFFF")
	roomdic.hset(roomid,"effect","bounce")
	roomdic.hset(roomid,"roomtoken",token)
	roomdic.hset(roomid,"bgblinkinterval",0)

	return jsonify({"roomid":roomid,"token":token})

	
@app.route("/api/<action>")
def api(action):
    no=request.args.get("no")
    token=request.args.get("token")
    roomid=request.args.get("roomid")
    if action=="settextcolor":
        color=request.args.get("color").replace("@","#")
        roomdic.hset(roomid,"textcolor",color)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"settextcolor","color":color}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done"
    elif action=="setbackgroundcolor":
        color=request.args.get("color").replace("@","#")
        roomdic.hset(roomid,"backgroundcolor",color)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"setbackgroundcolor","color":color}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done"    
    elif action=="settextsize":
        size=request.args.get("size")
        roomdic.hset(roomid,"textsize",size)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"settext","text":roomdic.hget(roomid,"text"),"size":float(roomdic.hget(roomid,"textsize"))}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done" 
    elif action=="settext":
        text=request.args.get("text")
        roomdic.hset(roomid,"text",text)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"settext","text":roomdic.hget(roomid,"text"),"size":float(roomdic.hget(roomid,"textsize"))}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done"         
    elif action=="setspeed":
        speed=request.args.get("speed")
        roomdic.hset(roomid,"speed",speed)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"setspeed","speed":abs(float(speed))}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"forceresync"}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done" 
    elif action=="setbgblinkinterval":
        interval=request.args.get("interval")
        roomdic.hset(roomid,"bgblinkinterval",interval)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"setbgblinkinterval","interval":interval}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done"         
    elif action=="seteffect":
        effect=request.args.get("effect")
        if not effect in ["normal","bounce"]:
        	return "invaild payload"
        roomdic.hset(roomid,"effect",effect)
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"seteffect","effect":effect}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"forceresync"}}))
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"updatepannelui"}}))
        return "done"     
    return "invaild method" 
def getuserlist(roomid):
	return userlist.lrange(roomid,0,-1)
def isroomempty(roomid):
	return len(getuserlist(roomid))==0
def ismaster(roomid,no):
    temp=getuserlist(roomid)
    return temp.index(str(no))==0
def getselfindex(roomid,no):
    temp=getuserlist(roomid)
    return temp.index(str(no))  
def getuserbefore(roomid,no,onlyone=True):
    temp=getuserlist(roomid)
    t=temp.index(str(no))-1
    if t<0:
        return[]
    if not onlyone:
        return temp[:t+1]
    return temp[t]  
def getuserafter(roomid,no,onlyone=True):
    temp=getuserlist(roomid)
    t=temp.index(str(no))+1
    if t>len(temp)-1:
        return[]
    if not onlyone:
        return temp[t:]
    return temp[t] 
def gettextwidth(no):
    result=devicedic.hget(no,"textwidth")
    return float(result) if result else 0
def getwidth(t):
    if isinstance(t,list):
        return [getwidth(i) for i in t]
    else:
        return float(devicedic.hget(t,"screenwidth"))
    
def getdevicewidthlist(roomid):
    return getwidth(getuserlist(roomid))
def getcanvaswidth(roomid):
    result=sum(getdevicewidthlist(roomid))
  #  assert result>0
    return result
def getwidthafteruser(roomid,no):
    return getwidth(getuserafter(roomid,no,onlyone=False))    
def getwidthbeforuser(roomid,no):
    return getwidth(getuserbefore(roomid,no,onlyone=False))
def getsafezone(roomid):
    return gettextwidth(getuserlist(roomid)[0])
def getcycledata(roomid,no):
    print(getwidthbeforuser(roomid,no),getwidthafteruser(roomid,no),no)
    rightwidth=sum(getwidthafteruser(roomid,no))
    leftwidth=sum(getwidthbeforuser(roomid,no))+getsafezone(roomid)
    print((leftwidth,rightwidth))
    return (leftwidth,rightwidth)
    
def get_pos_incavas(roomid,latency=0):#返回最左点(距左边缘,距右边缘)
    def get_towards():
        return "right" if towards>0 else "left" 
    def parseval():
        if valid_movement>=0:
          return (valid_movement-100,width-valid_movement)
        else:
          return (width+valid_movement-100,width-valid_movement)
    def effect_normal():
        valid_movement=total_movement%width
        return (width-valid_movement if speed<0 else valid_movement,speed)
    def effect_bounce():
        valid_movement=total_movement%width
        if temp%2==0:
          return(width-valid_movement if speed<0 else valid_movement,speed)
        else:
          return(width-valid_movement if speed>0 else valid_movement,-speed)
 
    now=time.time() * 1000
    speed=float(roomdic.hget(roomid,"speed"))
    ms_passed=now-float(roomdic.hget(roomid,"start_timestamp"))+latency
    movement_per_ms=60/1000*abs(speed)
    total_movement=movement_per_ms*ms_passed
    safezone=getsafezone(roomid)
    width=getcanvaswidth(roomid)+safezone
    temp=int(total_movement/width)
    #########################
    effectdic={"normal":effect_normal,"bounce":effect_bounce}
    valid_movement,towards=effectdic[roomdic.hget(roomid,"effect")]()
    #以下假设为跑马灯效果
    #valid_movement,towards=effect_normal()
    #以下假设为反弹效果
    #valid_movement,towards=effect_bounce()
    #return (parseval(),get_towards()) #debug only
    return (valid_movement-safezone if (valid_movement>=0) else width+valid_movement-safezone,get_towards(),temp)

def selfpos(roomid,no,pos=None):#返回弹幕在指定设备屏幕上的位置(以其屏幕大小为基准)
    px_before=sum(getwidthbeforuser(roomid,no))
    selfwidth=getdevicewidthlist(roomid)[getselfindex(roomid,no)]
    if not  pos:
      pos=get_pos_incavas(roomid)
    temp=pos[0]-px_before
    return (temp,pos[1])

""""
roomdic.hset(0,"text","abcd")
roomdic.hset(0,"speed",-10)
roomdic.hset(0,"start_timestamp",time.time() * 1000)
roomdic.hset(0,"musicbegin_timestamp",time.time() * 1000)
roomdic.hset(0,"textsize",400)
roomdic.hset(0,"textcolor","#006400")
roomdic.hset(0,"backgroundcolor","#FFFFFF")
roomdic.hset(0,"effect","bounce")
roomdic.hset(0,"roomtoken",KEY())
"""


def t(roomid):
    roundcount=-1
    while True:
      time.sleep(0.1)
      if not getuserlist(roomid): 
        continue
      result=get_pos_incavas(roomid)
      roomdic.publish("render_room_"+str(roomid),json.dumps(result)) 
      if roundcount!=result[-1]:
        print("onecycle")
        #userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"forceresync"}}))
        roundcount=result[-1]
        
      #print(get_pos_incavas(roomid))
#roomid=0      
#threading.Thread(target=t,args=(roomid,)).start()
def subscribe_pos(roomid,no):
                ps =roomdic.pubsub()
                ps.subscribe("render_room_"+str(roomid))    
                for item in ps.listen():
                  if item['type'] == 'message':
                    result=selfpos(roomid,no,pos=json.loads(item['data']))
                    #print(result)
                    ps.unsubscribe()
                    return result
def get_pos(roomid,no,latency=0):
        return selfpos(roomid,no,pos=get_pos_incavas(roomid,latency))
                

def broadcast_cycledata(roomid,no):
    for no in getuserlist(roomid):
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":no,"data":{"action":"setcycledata","data":getcycledata(roomid,no)}}))
def broadcast_selfindex(roomid,no):
    for no in getuserlist(roomid):
        userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":no,"data":{"action":"setindex","index":getselfindex(roomid,no)+1}}))


@app.route('/ws')
def feed(): 
    ws = request.environ.get("wsgi.websocket")
    latency=0
    while True:
        data = ws.receive()
        if not data:
            print("USER OFFLINE")
            abort(403)
        #print('Received: ' + data)
        data=json.loads(data)
        action=data["action"]
        
        if action=="statusreport":
            #if data["isrendering"]:
                if not data["current"]:
                	continue
                p=get_pos(roomid,no,-latency)
                if abs(data["current"]-p[0])>5:
                  print("误差过大", data["current"]-p[0])
                  ws.send(json.dumps({"action":"resync"}))
           # if (abs((time.time() * 1000-float(roomdic.hget(roomid,"musicbegin_timestamp")))/1000%data["soundduration"]-data["soundpos"]))>0.2:
           #     print("音乐不同步")
           #     ws.send(json.dumps({"action":"resyncmusic"}))   
        elif action=="reporttextwidth":
             devicedic.hset(no,"textwidth",data["textwidth"])
             broadcast_cycledata(roomid,no)
        
        elif action=="reportscreenwidth":
              devicedic.hset(no,"screenwidth",data["screenwidth"])
              broadcast_cycledata(roomid,no)
              userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"forceresync"}}))
        elif action=="syncmusic":
              ws.send(json.dumps({"action":"setmusicpos","pos":time.time() * 1000-float(roomdic.hget(roomid,"musicbegin_timestamp"))}))   
        elif action=="resync":
              temp=get_pos(roomid,no,latency)
              ws.send(json.dumps({"action":"setpos","towards":temp[1],"px":temp[0],"timestamp":time.time() * 1000}))
        elif action=="pong":
        	  latency=(time.time() * 1000-float(devicedic.hget(no,"lastping_timestamp")))/2
        	  error_in_ms=time.time() * 1000-(data["timestamp"]+latency)
        	  if abs(error_in_ms)>5:
        	  	 ws.send(json.dumps({"action":"settimestampoffset","offset":error_in_ms}))
        	  #print(latency)
        	  if latency>300:
        	    userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":no,"data":{"action":"showmsg","msg":"延迟过高("+str(latency)+"ms)"}}))
        	  
        	  
                
             
             
             
        elif action=="init":
            print(data)
            roomid=data["roomid"]
            roomtoken=data["roomtoken"]
            if not verify_roomtoken(roomid,roomtoken)["success"]:
            	ws.send(json.dumps({"action":"forcequit"}))
            	abort(403)
            userdic.incr("usercount")
            no=userdic.get("usercount")
            token=KEY()
            ws.send(json.dumps({"action":"setid","id":no,"token":token}))
            ws.send(json.dumps({"action":"setstarttimestamp","timestamp":roomdic.hget(roomid,"start_timestamp")}))
            ws.send(json.dumps({"action":"setspeed","speed":abs(float(roomdic.hget(roomid,"speed")))}))
            ws.send(json.dumps({"action":"setbgblinkinterval","interval":roomdic.hget(roomid,"bgblinkinterval")}))
            ws.send(json.dumps({"action":"settextcolor","color":roomdic.hget(roomid,"textcolor")}))
            ws.send(json.dumps({"action":"setbackgroundcolor","color":roomdic.hget(roomid,"backgroundcolor")}))
            ws.send(json.dumps({"action":"settext","text":roomdic.hget(roomid,"text"),"size":float(roomdic.hget(roomid,"textsize"))}))
            def run(ws):
                time.sleep(0.5)
                ws.send(json.dumps({"action":"forceresync"}))
            threading.Thread(target=run,args=(ws,)).start()
            ################################################
            def roombroadcast(ws,roomid):
                ps = userdic.pubsub()
                ps.subscribe("room_"+str(roomid),no)
                for item in ps.listen():
                  if item['type'] == 'message':
                    if item['data']=="OFFLINE":
                        ps.unsubscribe()
                        print("广播监听线程退出")
                        return
                    if item['channel']=="room_"+str(roomid):
                        msg=json.loads(item['data'])
                        if msg["action"]=="totarget":
                          if no in msg["target"]  or msg["target"]=="all":
                            ws.send(json.dumps(msg['data']))
            def heartbeat(ws,no):
              while True:   
                 try:
                      ws.send(json.dumps({"action":"ping"}))
                      devicedic.hset(no,"lastping_timestamp",time.time() * 1000)
                      time.sleep(0.5)
                 except:
                    import traceback
                    #traceback.print_exc()
                    print("守护心跳线程退出",no)
                    userdic.publish(no,"OFFLINE")
                    userdic.delete(no)
                    userlist.lrem(roomid,0,no)
                    devicedic.delete(no)
                    broadcast_cycledata(roomid,no)
                    broadcast_selfindex(roomid,no)
                    userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"forceresync"}}))  
                    userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"showmsg","msg":"一个用户下线"}}))
                    if isroomempty(roomid):
                    	roomdic.delete(roomid)
                    	print("房间已被删除")
                    return
            threading.Thread(target=roombroadcast,args=(ws,roomid)).start()
            threading.Thread(target=heartbeat,args=(ws,no)).start()
            devicedic.hset(no,"ua",data["ua"])
            devicedic.hset(no,"token",token)
            devicedic.hset(no,"screenwidth",data["screenwidth"])
            userlist.rpush(roomid,no)
            ws.send(json.dumps({"action":"setindex","index":getselfindex(roomid,no)+1}))
            userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"forceresync"}}))
            userdic.publish("room_"+str(roomid),json.dumps({"action":"totarget","target":"all","data":{"action":"showmsg","msg":"一个新用户成功加入"}}))
            
      
             
if __name__ == '__main__':
    print("warning：启动方式错误！！！！！！！！！！！！！！！！！！！！！")
    print("warning：启动方式错误！！！！！！！！！！！！！！！！！！！！！")
    print("warning：启动方式错误！！！！！！！！！！！！！！！！！！！！！")
    print("使用： gunicorn3 -k geventwebsocket.gunicorn.workers.GeventWebSocketWorker website:app -b 0.0.0.0:6000")              
    os._exit(0)
    print("warning：启动方式错误")        
    
   
    


