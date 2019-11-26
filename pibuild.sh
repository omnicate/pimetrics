TARGET=pimetrics_arm

env GOOS=linux GOARCH=arm go build -o ./bin/$TARGET
if [ $? == 0 ]
then
    scp ./bin/$TARGET wg2@192.168.1.110:/home/wg2/$TARGET
    scp -r ./static wg2@192.168.1.110:/home/wg2/
fi