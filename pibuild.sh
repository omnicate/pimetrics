TARGET=pimetrics_arm

PI=$1

env GOOS=linux GOARCH=arm go build -o ./bin/$TARGET
if [ $PI == "pi3" ]
then
    scp ./bin/$TARGET ubuntu@192.168.1.48:/home/ubuntu/$TARGET
    scp ./config.yaml ubuntu@192.168.1.48:/home/ubuntu/config.yaml
    scp -r ./web ubuntu@192.168.1.48:/home/ubuntu/
elif [ $PI == "pi4" ]
then
    scp ./bin/$TARGET ubuntu@192.168.1.108:/home/ubuntu/$TARGET
    scp ./config.yaml ubuntu@192.168.1.108:/home/ubuntu/config.yaml
    scp -r ./web ubuntu@192.168.1.108:/home/ubuntu/
fi