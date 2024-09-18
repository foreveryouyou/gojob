#!/usr/bin/env bash

#########################################
# 在终端执行该文件, 参数传要执行的函数名即可
# 如: ./run.sh usage
#########################################

cwd=$(pwd)
NC='\033[0m'
C_RED='\033[1;31m'
C_GREEN='\033[1;32m'
C_YELLOW='\033[1;33m'
C_BLUE='\033[1;34m'
C_PURPLE='\033[1;35m'
C_SKY='\033[36m'

# 使用说明，用来提示输入参数
usage() {
    echo -e "${C_BLUE}Usage:${NC}\n ${C_GREEN}./run.sh${NC} [command]\n"
    echo -e "${C_BLUE}Available Commands:${NC}"
    echo -e " ${C_GREEN} example     ${NC} 运行示例代码"
    echo -e " ${C_GREEN} docker      ${NC} 启动 docker 环境"
    echo
}

example() {
    go run example/main.go $@
}

docker() {
    cd ./docker && docker-compose up $@
}

###############################################################################

if [ $# -eq 0 ]; then
    usage
    echo -e "${C_YELLOW}请输入要执行的功能${NC}"
    exit 1
fi
# 判断函数是否存在
func_name=$1
type $func_name >/dev/null 2>&1
if [ $? -ne 0 ]; then
    if [ $func_name != '-h' ]; then
        echo -e "${C_RED}不支持的功能: ${func_name}${NC}"
        echo
    fi
    usage
    exit 1
fi
# 执行指定函数
shift
$func_name "$@"
