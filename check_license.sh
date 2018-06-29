
echo "Checking Go files for license headers ..." 
 missing=`find . -name "*.go" | grep -v build/ | grep -v vendor/ | grep -v statistics/ | grep -v cloudflare/ | grep -v ".pb.go" | grep -v "_test.go" | xargs grep -l -i -L "License"` 
 if [ -z "$missing" ]; then 
 echo "All go files have license headers" 
 exit 0 
 fi 
 echo "The following files are missing license headers:" 
 echo "$missing" 
 exit 1 