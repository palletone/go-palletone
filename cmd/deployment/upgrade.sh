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
			length=${#table}
			num=$[$length-2]
			table=${table:1:$num}
			#echo $table
			#grep -n '\[Contract\]'  ptn-config.toml.new | head -1 | cut -d ":" -f 1
			table="\[$table\]"
	echo "==============================="$2 $table
			num=`grep -n $table  $FileName | head -1 | cut -d ":" -f 1`
	echo "====add====="$file $table $content $num
	#sed '/"$table"/a\"$content"' $file >> $file
	#num=`grep -n "$table"  $FileName | head -1 | cut -d ":" -f 1`
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


function addtable(){
	file=$1
	table=$2
	echo "====addtable====="$file $table
	echo $table>>$file
	echo -e "\n">> $file
	echo -e "\n">> $file
	echo -e "\n">> $file

}

function deltable(){
	file=$1
	table=$2
	line=$3
        if [[ "$table" != "$jury" ]] && [[ "$table" != "$mediator" ]]
        then
		echo "====deltable====="$file $table
                sed -i ${line}d $file
		#sed -i "/$table/d" $file
        fi

}

function while_read_bottm(){
	line=-1
	srctable=""
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
					addcontent $3 $srctable $content
				fi
				#del key in filename
				if [[ "$4" == "2" ]]
				then
					delcontent $3 $srctable $content $key
				fi
			fi

		else
			srctable=$content
    			table=$content
			length=${#table}
			num=$[$length-2]
			table=${table:1:$num}
			#echo $table
			#grep -n '\[Contract\]'  ptn-config.toml.new | head -1 | cut -d ":" -f 1
			table="\[$table\]"
			flag=`grep -n $table  $2 | head -1 | cut -d ":" -f 1`
			echo $table "grep" $2 "flag:"$flag
			if [[ "$flag" == "" ]]
			then
				#add key in filename
				if [[ "$4" == "1" ]]
				then
					addtable $3 $content
				fi
				#del key in filename
				if [[ "$4" == "2" ]]
				then
					line=`grep -n $table  $3 | head -1 | cut -d ":" -f 1`
					deltable $3 $content $line
				fi
			fi
		fi
	done < $1
}
cp $FileNameOld $FileName 
#$1 compare to $2  add $3
while_read_bottm $FileNameNew $FileName $FileName 1


while_read_bottm $FileNameOld $FileNameNew $FileName 2




