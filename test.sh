#!/bin/sh
source ./colorlog.sh
redlog "哈哈哈test"

if [ -n ""  ];then
	echo "Get the gift Num "
fi

if [ "`netstat -lnt | grep 8080 |awk -F '[:]' '{print $4}'`" -eq 8080 ];then
	echo "8080 is running"
 else
	echo redlog "$port is stoped"
fi

op=(8080 9090)	
for i in ${op[@]}
do
	if [ "`netstat -lnt | grep ${i} | awk -F '[:]' '{print $4}'`" -eq ${i} ];then
      greenlog "${i} is running"
    else
      echo redlog "${i} is stoped"
    fi

done



