# go-ping

go で ping コマンドを実装。SoftWareDesign 2021 年 5 月号

- ICMP パケットは ping コマンドのためだけでなく、インターネット・プロトコルのデータグラム処理における誤りの通知や、通信に関する情報の通知などのために使用される。
- ping コマンドでは ICMP パケットのうち、「ICMP Echo Request」と「ICMP Echoh Reply」を利用して疎通確認を行う。
