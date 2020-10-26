# pimetrics
Pi metrics for calling and network testing


## SMS by curling

```
curl -X POST -H "Content-Type: application/json" -d '{"number":"+4790300231", "text":"SMS from curl"}' http://192.168.1.110:8080/send_sms
```

## Service updating and discovery

Pimetrics uses the following bucket `swt-telco-lab-dev` in the dev env for pulling
config to the raspberry pis.

The config looks like:

```yaml
device-id:
  tenent: vimla
  msisdn: 46123456789
  target: arm64
  sw_version: 1.0.0
  config: |
    some_yaml: yaml
```

This way each pi would get it's own config and would be able to dicover other devices
in the lab.