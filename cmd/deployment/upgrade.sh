#!/bin/bash

FileName="./ptn-config.toml"
FileNameOld="./ptn-config.toml.old"
FileNameNew="./ptn-config.toml.new"

jury="[[Jury.Accounts]]"
mediator="[[MediatorPlugin.Mediators]]"

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

function delcontent(){
	file=$1
	table=$2
	content=$3
	key=$4
	#if [[ "$table" != "$jury" ]]
	if [[ "$table" != "$jury" ]] && [[ "$table" != "$mediator" ]] 
	then
		echo "====del====="$file $table $content
		#sed -i '/"$key"/d' $file
		sed -i "/$key/d" $file
	fi
}

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
			flag=`grep -n "$key"  $2 | head -1 | cut -d ":" -f 1`
   			#echo "line:"$line "table:"$table "content:"$content "包含:" $key "flag:"$flag

			if [[ "$flag" == "" ]]
			then
				#add key in filename
				if [[ "$4" == "1" ]]
				then
					addcontent $3 $table $content
				fi
				#del key in filename
				if [[ "$4" == "2" ]]
				then
					delcontent $3 $table $content $key
				fi
			fi

		else
    			table=$content
		fi
	done < $1
}
cp $FileNameOld $FileName 
#$1 compare to $2  add $3
while_read_bottm $FileNameNew $FileName $FileName 1


while_read_bottm $FileNameOld $FileNameNew $FileName 2




