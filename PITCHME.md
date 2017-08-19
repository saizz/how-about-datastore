### How about datastore

GCPUG Shonan vol.16 feat.Datastore

---

### お品書き

- Case1 トランザクション同時実行の制約
- Case2 トランザクション内、Entity Groupの制約
- Case3 トランザクションの有効期限
- Case4 task queue登録後にtransactionのroll back

---

### 検証プログラムのデプロイ

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

では、次にconcurrent=2, child=1とした場合、どうなるか

---

https://xxx.appspot.com/case1?concurrent=2&child1

putのログは2つ出力されるはずだが、3つ出力されている・・・

![result-log](https://docs.google.com/drawings/d/14oYjNwR2okmw1Q34UT-Bh_LRgq3z-PpS1W6cVcwdmUA/pub?w=586&h=390)

---

思わず、traceを見てみる

![result-trace](https://docs.google.com/drawings/d/1_GhYRHDJr0Xcq9NJU_cG4pOmlZ37tb5H_oeYB54bwkc/pub?w=865&h=425)

---

GAE/Goでは、Appengine SDK側で、デフォルト2回のリトライをするようになっている

![appnegine-sdk](https://docs.google.com/drawings/d/1QMBrcEXLt0VWRQadbYgP_8oIxO_dEbH6q_ZdqcgYGIQ/pub?w=847&h=350)

---

ドキュメントにも書いてある

https://cloud.google.com/appengine/docs/standard/go/datastore/transactions

![trx-document](https://docs.google.com/drawings/d/1U9EiPEuMODIjqBw61uhAQ9V7GjONN0PmWvdmjTe2pGc/pub?w=902&h=213)

---

結果は

![result-sheet](https://docs.google.com/drawings/d/1Fxs4kvROo8andi_fsL3i6BPn3SJ2p06AeTKJv0yGHo4/pub?w=686&h=272)

---

concurrent=10にしたら、1回目で9 transactionは失敗して、2回目で8 transactionが失敗して、となりそうだが・・・

---

https://xxx.appspot.com/case1?concurrent=10&child1
![result-trace-10_1](https://docs.google.com/drawings/d/1OmewpqJV7tqgVxTSAtlelMRlM8hhgz79w2NrVlj6gFA/pub?w=960&h=720)

---

![result-trace-10_2](https://docs.google.com/drawings/d/199DKUgqeiIlC5KAqzvt0FkEYZakoYZVoRlcG8FNUNXQ/pub?w=960&h=720)

---

- transactionの同時実行で、commit失敗が戻ってくるタイミングが微妙にずれるので、2回目のtrasactionは、予想より成功している
- つまり、同時でなければ、順番にtransactionは成功する

---

### Case2 トランザクション内、Entity Groupの制約

---

https://cloud.google.com/appengine/docs/standard/go/datastore/entities

![eg-doc](https://docs.google.com/drawings/d/1trTagwXOTk4cuThG6Bg1Pyw8dS7ASx3T-yz9Qnn8EdY/pub?w=902&h=273)

---

1 transaction内でputするEntity Groupの数(parentの数)を変えていった場合、制約がどうあらわれるか?

---

図にするとこんな感じ

![model](https://docs.google.com/drawings/d/1Mk9-R8KuZcjA77YwzYDMdkVtG_6ctfTikcSSHdPmOkg/pub?w=847&h=485)

---

concurrent=1

![eg-sheet](https://docs.google.com/drawings/d/1rjJ8vLP6-Q8XZL5ksH59Wzt4VE51j2ElJlqcxtXbubc/pub?w=475&h=265)

---

parent=1ならCase1と変わらない。parent=2以上の場合のどうなるか

---

結果は

![result](https://docs.google.com/drawings/d/1AyiY_7itdPuWYl6xZb8zqx6Kgs0w2h6dMpqblxFNFnA/pub?w=475&h=265)

---

https://xxx.appspot.com/case2?concurrent=1&parent=26&child=1

![log](https://docs.google.com/drawings/d/1ZsP-oTQGheoxlFwbZaurN9ppYrdMPonZSuhSe3T2sqU/pub?w=730&h=102)

API error 1 (datastore_v3: BAD_REQUEST): operating on too many entity groups in a single transaction.

---

single group transactionとcross group transactionは、仕組み自体が異なる様子

![go-doc](https://docs.google.com/drawings/d/19h1Z6_dkWwRtK1kAbt695d-C7W67kfCpNFDqylw2Pdw/pub?w=677&h=293)

---

- 1 transaction内で更新可能なEntity Groupの上限は、仕組み上、25という上限がありそう
- cross group transactionを有効にいていないと、2の場合でもエラーを返される
- cross group transactionが有効で、実際に更新するEntity Groupが1つであっても、それはエラーにはしていない

---

### Case3 トランザクションの有効期限

---

![expired-doc](https://docs.google.com/drawings/d/1iQ9ngAy639TtryI7u2ulmpXfAZWSum9S7CPWa585orQ/pub?w=791&h=355)

---

- task queueでtransactionを実行
- transaction内でsleepで時間を潰す
- sleep時間を60秒前後でどうなるか

---

図にすると、こんな感じ

![model](https://docs.google.com/drawings/d/18hrdJY6hwJfBRNT1dOiuSXuMfADBefS78VivP83HDNc/pub?w=470&h=278)

---

![sheet](https://docs.google.com/drawings/d/1IA1Ok6fJBwRbmeR-I-fjJ761YtHma5flaS0kjxN-OGw/pub?w=304&h=307)

- nを70にした場合、どうなるか

---

https://xxx.appspot.com/case3?n=70

![log](https://docs.google.com/drawings/d/1EG3h6ys5buhNeFqITgoKFxRj66JwufjAEfZ3yjOMh_s/pub?w=777&h=397)

---

結果は

![result](https://docs.google.com/drawings/d/1i9MsFdM6wA3W-O6o-tV7OkebSfvMMMdU-AbUoyZvIIQ/pub?w=304&h=307)

---

- 60秒たった時点で、transactionが打ち切りになるわけでない
- 70秒たって、RunInTransactionからエラーが返る

---

### Case4 task queue登録後にtransactionのroll back

---

![taskqueue-doc](https://docs.google.com/drawings/d/1vmSriP-IOrwaV-lio939eCGu3s-ac6yXhDWTtg84AQI/pub?w=899&h=196)

---

- transaction内でtask queueの登録を行い、その後にroll backされた場合、task queueは実行されないのか?

---

図にすると、こんな感じ

![model](https://docs.google.com/drawings/d/1CByu9tYkfCOW-sQ8sZL_TzNZaMocEMYaSp_fdv5SOcM/pub?w=553&h=233)

---

![sheet](https://docs.google.com/drawings/d/1NAqhT5mb6QpuxEs8-H3RL3Vvr2Izzcx_CL3y_3uqCrM/pub?w=252&h=177)

---

結果は

![result](https://docs.google.com/drawings/d/18HD42Xua7FeDFdjPj_fRTkE3KySVbbhrIkbqJMtW_Co/pub?w=807&h=463)

---

- transactionがcommit成功した時のみ、task queueの呼び出しがされる
