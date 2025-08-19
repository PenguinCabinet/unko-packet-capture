# 💩 unko packet capturer
![img](/image/gif1.gif)

パケットをうんちにして表示するパケットキャプチャです。

送信パケットは左から右、受信パケットは右から左に流れます。

## 使い方
```
unko-packet-capture.exe pd
```
```
unko-packet-capture.exe nd
```
上記二つのコマンドを実行し、キャプチャーするネットワークアダプターを選んでください。

```
unko-packet-capture.exe -pdev "<pdで表示したアダプターのName>"  -ndev "<ndで表示したアダプターのName>"
```
