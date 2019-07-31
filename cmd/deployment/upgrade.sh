#!/bin/bash

FileName="./ptn-config.toml"
FileNameOld="./ptn-config.toml.old"
FileNameNew="./ptn-config.toml.new"

#jury="[[Jury.Accounts]]"
#mediator="[[MediatorPlugin.Mediators]]"
jury="Jury"
mediator="Mediator"

function addcontent(){
	file=$1
	table=$2
	content=$3
	length=${#table}
	num=$[$length-2]
	table=${table:1:$num}
	table="\[$table\]"
	num=`grep -n $table  $FileName | head -1 | cut -d ":" -f 1`
	echo "====add====="$file $table $content $num
	num=$[$num+1]
	sed -i "${num}i $content" $file
}

function delcontent(){
	file=$1
	table=$2
	content=$3
	key=$4
	#if [[ "$table" != "$jury" ]] && [[ "$table" != "$mediator" ]] 
	if [[ $table != *$jury* ]] && [[ $table != *$mediator* ]]
	then
		echo "====del====="$file $table $content
		sed -i "/$key/d" $file
	fi
}


function addtable(){
	file=$1
	table=$2
	if [[ $table != *$jury* ]] && [[ $table != *$mediator* ]]
	then
		echo "====addtable====="$file $table
		echo $table>>$file
		echo -e "\n">> $file
		echo -e "\n">> $file
		echo -e "\n">> $file
	fi

}

function deltable(){
	file=$1
	table=$2
	line=$3
        if [[ $table != *$jury* ]] && [[ $table != *$mediator* ]]
        then
		echo "====deltable====="$file $table
                sed -i ${line}d $file
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
			table="\[$table\]"
			flag=`grep -n $table  $2 | head -1 | cut -d ":" -f 1`
			#echo $table "grep" $2 "flag:"$flag
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

Path="mainnet"
#Path="testnet"

rm -rf temp
mkdir temp
cd temp
#wget release package
wget $1
tar xzvf *.tar.gz
cd $Path
dumcpconfig=`./gptn dumpconfig`
echo $dumpconfig
cp ptn-config.toml ../ptn-config.toml.new
cd ../
cp ../ptn-config.toml ptn-config.toml.old

cp $FileNameOld $FileName 
#$1 compare to $2  add $3
while_read_bottm $FileNameNew $FileName $FileName 1

#$1 compare to $2 del $3
while_read_bottm $FileNameOld $FileNameNew $FileName 2


mv ../ptn-config.toml ../ptn-config.toml.bak
cp ./ptn-config.toml ../ptn-config.toml

mv ../gptn ../gptn.bak
cp $Path/gptn ../gptn
