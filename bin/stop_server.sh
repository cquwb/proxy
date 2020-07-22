kill -9 $(ps -ef|grep proxyserver|grep -v grep|awk '{print $2}')
