#!/bin/bash

FileName="./ptn-config.toml"
FileNameOld="./ptn-config.toml.old"
FileNameNew="./ptn-config.toml.new"


function addcontent(){
	file=$1
	table=$2
	content=$3
	echo "====add====="$file $table $content
	#sed '/"$table"/a\"$content"' $file >> $file
	num=`grep -n "$table"  $FileName | head -1 | cut -d ":" -f 1`
	num=$[$num+1]
	sed -i "${num}i $content" $file
}


function while_read_bottm_new(){
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
   			echo "line:"$line "table:"$table "content:"$content "包含:" $key "flag:"$flag

			if [[ "$flag" == "" ]]
			then
			#add key in filename
				addcontent $FileName $table $content
			fi

		else
    			table=$content
		fi
	done < $FileNameNew
}
cp $FileNameOld $FileName 
while_read_bottm_new




#https://www.jb51.net/article/156339.htm


