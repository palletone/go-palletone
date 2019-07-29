#!/bin/bash

FileName="./ptn-config.toml"
FileNameOld="./ptn-config.toml.old"
FileNameNew="./ptn-config.toml.new"

function while_read_bottm(){
	line=-1
	table=""
	while read content
	do
		line=$[$line+1]
		
		if [ "$content" =  "" ]
		then
			continue
		fi
		
		content=`echo $content | sed 's/[[:space:]]//g'`
		#content=`echo $content | sed -e 's/^[ \t]*//g'`
		#echo $content
		result=$(echo $content | grep "=")
		if [[ "$result" != "" ]]
		then
			subIndex=`echo $content | grep -b -o "=" | cut -d: -f1`
			key=${content:0:$subIndex}
			key=`echo $key | sed -e 's/^[ \t]*//g'`

			#compare to filenameold and add new key
			flag=`grep -n "$key"  $FileName | head -1 | cut -d ":" -f 1`

			if [[ "$flag" != "" ]]
			then
			#add key in filename
				
			fi

   			echo "line:"$line "table:"$table "content:"$content "包含:" $key "flag:"$flag
		else
    			table=$content
		fi
	done < $FileNameNew
}
cp $FileNameOld $FileName 
while_read_bottm




#https://www.jb51.net/article/156339.htm


