while read line; do
name=`echo $line|awk -F '=' '{print $1}'`
value=`echo $line|awk -F '=' '{print $2}'`
case $name in
"IPCPath")
echo "IPCPath:"+$value
;;
"pwd")
pwd=$value
;;
"permission")
permission=$value
;;
*)
;;
esac
done < ptn-config.toml 

#while read line; do
#name=`echo $line|awk -F '=' '{print $1}'`
#value=`echo $line|awk -F '=' '{print $2}'`
#echo "name:"+ $name
#echo "value:"+$value
#done < ptn-config.toml 
