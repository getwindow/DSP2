#!/bin/sh
#输出代用颜色的内容
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

con=$(redlog "世界这么大我想去走走")
echo $con
echo $con
