#!/bin/sh
portArr=(8080 8081 8082 8083 9090)

RED_COLOR='\E[1;31m'
GREEN_COLOR='\E[1;32m'
YELLOW_COLOR='\E[1;33m'
BLUE_COLOR='\E[1;34m'
RES='\E[0m'
 
redlog(){
   echo -e  ${RED_COLOR}$1${RES}
 }
  
greenlog(){
      echo -e ${GREEN_COLOR}$1${RES}
  }

yellowlog(){
     echo -e ${YELLOW_COLOR}$1${RES}
  }
  
bulelog(){
     echo -e ${BLUE_COLOR}$1${RES}
  }



#开动进程
stop(){
	if [ ${#1} -gt 0 ];then
		 port=$(lsof -i:$1 | awk '/^main/{print $2}')
          kill $port
          if [ $? -eq 0 ] ;then
            redlog "端口:$1 关闭"
          fi

	   return 
	fi

	for i in ${portArr[@]}
  	do
	    portone=$(lsof -i:$i | awk '/^main/{print $2}')
	    kill $portone
	    if [ $? -eq 0 ]	;then 
		  redlog "端口:$i 关闭"
	    fi
	done

}

#关闭当前的进程
start(){
      if [ ${#1} -gt 0 ];then
              echo "The length is more than one "
	  	      #port=$(lsof -i:$i | awk '/^main/{print $2}')			
		     nohup ./main -port ":$1" >/dev/null 2>&1 &
             if [ $? -eq 0 ];then
              greenlog "端口:$1启动成功"
             fi

             return
       fi
 
      for i in ${portArr[@]}
      do
        port=$(lsof -i:$i | awk '/^main/{print $2}')
		nohup ./main -port ":${i}" >/dev/null 2>&1 &
        if [ $? -eq 0 ];then
			greenlog "端口:${i}启动成功"
		fi
     #  if [ ${#1} -gt 0 ];then return fi;
     done

}

#重启当前进程
restart(){
	stop $1 #当前方法的第二个参数
	start $1
}



#检查当前端口
rtest(){
 
 for i in ${portArr[@]}
  do
      if [ "`netstat -lnt | grep ${i} | awk -F '[:]' '{print $4}'`" -eq ${i} ];then
        greenlog "${i} is running"
      else
        redlog "${i} is stoped"
      fi 
  done

}


case "$1" in
    start)
       start $2
    ;;
    stop)
	   stop $2
    ;;
  restart)
	  restart $2
    ;;
   test)
	  rtest
	;;
esac
