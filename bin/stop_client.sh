kill -9 $(ps -ef|grep proxyclient|grep -v grep|awk '{print $2}')
