TARGET=pimetrics_arm

env GOOS=linux GOARCH=arm go build -o ./bin/$TARGET
if [ $? == 0 ]
then
    scp ./bin/$TARGET ubuntu@192.168.1.108:/home/ubuntu/$TARGET
    scp -r ./static ubuntu@192.168.1.108:/home/ubuntu/
fi