relativeFile=$1
lineNumber=$2
line=0
funcName=""
while :
do
  cursol=`expr $lineNumber - $line`
  if [ "$cursol" = "0" ]; then
    break
  fi

  strs=`sed -n ${cursol}p $relativeFile | grep func | grep testing.T`

  if [ "$strs" != "" ]; then
    funcName=$strs
    break
  fi

  line=`expr $line + 1`
done
funcName=`echo $funcName | cut -d " " -f 2 | cut -d "(" -f 1`
echo "Testing function: $funcName"
go test -run $funcName
