# pimetrics
Pi metrics for calling and network testing


## SMS by curling

```
curl -X POST -H "Content-Type: application/json" -d '{"number":"+4790300231", "text":"SMS from curl"}' http://192.168.1.110:8080/send_sms
```