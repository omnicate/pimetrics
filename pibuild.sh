TARGET=pimetrics_arm

env GOOS=linux GOARCH=arm go build -o ./bin/$TARGET
if [ $? == 0 ] || [ $1 == "pi3" ]
then
    scp ./bin/$TARGET ubuntu@192.168.1.48:/home/ubuntu/$TARGET
    scp ./config.yaml ubuntu@192.168.1.48:/home/ubuntu/config.yaml
    scp -r ./web ubuntu@192.168.1.48:/home/ubuntu/
elif [ $1 == "pi4" ]
then
    scp ./bin/$TARGET ubuntu@192.168.1.108:/home/ubuntu/$TARGET
    scp ./config.yaml ubuntu@192.168.1.108:/home/ubuntu/config.yaml
    scp -r ./web ubuntu@192.168.1.108:/home/ubuntu/
fi