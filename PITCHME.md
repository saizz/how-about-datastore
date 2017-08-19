### How about datastore

GCPUG Shonan vol.16 feat.Datastore

---

### お品書き

- Case1 トランザクション同時実行の制約
- Case2 トランザクション内、Entity Groupの制約
- Case3 トランザクションの有効期限
- Case4 task queueはroll back対象？

---

### 検証プログラムのデプロイデプロイ

```
# in cloud shell
$ git clone https://github.com/saizz/how-about-datastore
$ cd how-about-datastore
$ export GOPATH=$(pwd)

# change application in app.yaml
$ vi src/backend/app.yaml

# deploy
$ gcloud app create --region=asia-northeast1 --project=xxx
$ goapp deploy src/backend
```
---

### Case1 トランザクション同時実行の制約

---

- transactionを使うと一つのEntity Groupについて1回/秒の制約とよく言われるが具体的にどのような制約なのか?
- 10回/秒ぐらいはいけると聴いたこともある
- 実際にちょっとしたプログラムを書いて検証してみた

---
https://cloud.google.com/appengine/docs/standard/go/datastore/transactions

![ds-trx-001](https://docs.google.com/drawings/d/1lmCWEhpPO7FD5oUIJ26K3mkhj_lqwCutSy4QkSEW5TU/pub?w=910&h=316)

---

https://cloud.google.com/appengine/docs/standard/go/datastore/entities

![ds-trx-002](https://docs.google.com/drawings/d/1grP6O60ewGkLasIusR2_4nHeBhFLgLLbA8O6QR0L21U/pub?w=914&h=272)

---

https://cloud.google.com/appengine/docs/standard/go/datastore/structuring_for_strong_consistency

![ds-trx-003.png](https://docs.google.com/drawings/d/1Bw8iQ-DSE9d0_oVLm5dZlJDcamO5-KbsVoSboI1XEOw/pub?w=897&h=245)

---

- 同じEntity Groupへの更新をtransactionで行う
- 1 transaction内でputする回数を変えてみる(childの数)
- transactionの同時実行数を変えてみる(concurrentの数)
- childとconcurrentを変更した場合、制約がどのように現れるか
- putの結果をlogに出力

---

図にするとこんな感じ

![model](https://docs.google.com/drawings/d/15btUDt7gwEp3QBR_wVpnhbmfagla2PC9jn1phr2XR0o/pub?w=779&amp;h=358)

---

- concurrent=1の場合は、同時には1 transactionなので特に制約に引っかからない
![sheet](https://docs.google.com/drawings/d/1Q_4YEsBFD2_lSUoKqneKklzA2YRCtn_9dYZn3ECqmtg/pub?w=802&h=267)

---

では、次にconcurrent=2, child=2とした場合、どうなるか

---

https://xxx.appspot.com/case1?concurrent=2&child1

同時2なので、putのログは2つ出力されるはずだが、3つ出力されている・・・

![result-log](https://docs.google.com/drawings/d/14oYjNwR2okmw1Q34UT-Bh_LRgq3z-PpS1W6cVcwdmUA/pub?w=586&h=390)

---

思わず、traceを見てみる
![result-trace](https://docs.google.com/drawings/d/1_GhYRHDJr0Xcq9NJU_cG4pOmlZ37tb5H_oeYB54bwkc/pub?w=865&h=425)
