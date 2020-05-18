# Nezha
哪吒: 通过插件化灵活组合业务流程平台

# 设计目标
可以方便的接入各种数据源，中间可以加入自定义插件化的处理，再将处理结果输出到任意数据仓库中

# 架构设计

### 核心概念三点：

#### 1. Component(组件)接口

组件概念: 基础资源为之后的Handler方法提供依赖资源，注入依赖的根据。

#### 2. Processor(处理器)方法

处理器概念: 具体的业务逻辑的实现单元，Processor的类型必须是一个方法，入参/出参(error例外)被限制为不限个数的结构体/结构体指针类型
入参: 只接受结构体/结构体指针类型，非上述类型会返回 fmt.Errorf("Value not found for type %v", ft)
出参: 只接受结构体/结构体指针类型，非上述类型会被丢弃。如果返回值中有error类型并且非nil会将其返回

#### 3. Plugin接口

插件概念: 任意实现上述Processor方法或者Component接口的可做为插件进行集成


### How to use

例如我们现在有一个需求，从kafka消费的消息是一片文章，需要将它拆分成句子，然后将句子分别写入es和新的kafka


#### Step 1 准备流程中需要的component资源
component的实现可以参考 github.com/shima-park/nezha/pkg/component/{es, kafka}下相关实现。只考虑如何实现业务Processor的编写

已知的通过上述component可以获得以下对象，并已注入容器
```
"KafkaNewtonArticleConsumer" :  *cluster.Consumer
"ESNewtonDailyClient" :         *elastic.Client
"ESNewtonIssueClient" :         *elastic.Client
"KafkaNewtonSentenceProducer" : sarama.SyncProducer
```

#### Step 2 编写读取kafka消息的处理方法

``` go
// 作为业务逻辑开发时，用以下代码注册
processor.Register("ReadNewtonArticleMessageFromKafka", func(config string) (processor.Processor, error) {
    return ReadNewtonArticleMessageFromKafka, nil
})

/* 作为插件时，用以下代码注册
var Bundle = plugin.Bundle(
	processor.Plugin("ReadNewtonArticleMessageFromKafka", func(config string) (processor.Processor, error) {
        return ReadNewtonArticleMessageFromKafka, nil
    }),
)
*/

type Request struct{
    // 注入上下文context.Context对象
    Ctx                        context.Context `inject:"Ctx"`
    // 注入kafka消费者KafkaNewtonArticleConsumer
    KafkaNewtonArticleConsumer *cluster.Consumer `inject:"KafkaNewtonArticleConsumer"`
}

type Response struct{
    // 返回KafkaNewtonArticleMessage将其注入容器
    KafkaNewtonArticleMessage *sarama.ConsumerMessage `inject:"KafkaNewtonArticleMessage"`
}

func ReadNewtonArticleMessageFromKafka(r *Request) (res *Response, err error) {
    var msg *sarama.ConsumerMessage
    select {
        case <- r.Ctx.Done():
            return nil, nil
        case m, ok := <- r.KafkaNewtonArticleConsumer.Messages():
            if ok {
                msg = m
            }
    }

    return &Response{NewtonArticleMessage: msg},nil
}
```

#### Step 3 定义公用的数据结构 package, 用来在不同的plugin中共享数据结构


``` go
package proto

type NewtonArticleMessage struct {
    Allow      bool   `json:"allow_empty_date,omitempty"`
    PubDate    string `json:"pub_date,omitempty"`
    Title      string `json:"title,omitempty"`
    Url        string `json:"url,omitempty"`
    Content    string `json:"content,omitempty"`
    BundleKey  string `json:"bundle_key,omitempty"`
    Domain     string `json:"domain,omitempty"`
    SourceName string `json:"source,omitempty"`
    Html       string `json:"html,omitempty"`
    Publisher  string `json:"publisher,omitempty"`
}

type NewtonSetenceObject struct{
    Sentence  string `json:"sentence"`
	BundleKey string `json:"bundle_key"`
	PubDate   string `json:"pub_date"`
}
```

#### Step 4 编写解码数据为对应结构体方法

``` go
import "proto" // 导入共享的数据结构契约包
// import ...

// 作为业务逻辑开发时，用以下代码注册
processor.Register("DecodeNewtonArticleMessage", func(config string) (processor.Processor, error) {
        return DecodeNewtonArticleMessage, nil
})

/* 作为插件时，用以下代码注册
var Bundle = plugin.Bundle(
	processor.Plugin("DecodeNewtonArticleMessage", func(config string) (processor.Processor, error) {
        return DecodeNewtonArticleMessage, nil
    }),
)
*/

type Request struct{
    // 注入从消费kafka得来的NewtonArticleMessage
    NewtonArticleMessage *sarama.ConsumerMessage `inject:"NewtonArticleMessage"`
}

type Response struct{
    // 返回的NewtonArticleMessage对象会被注入到容器中，来提供给后面的流程使用
    NewtonArticleMessage *NewtonArticleMessage `inject:"NewtonArticleMessage"`
}

func DecodeNewtonArticleMessage(r *Request) (*Response, error) {
    var s proto.NewtonArticleMessage
    err := json.Unmarshal(r.NewtonArticleMessage.Value, &s)
    if err != nil {
        return nil, err
    }

    return &Response{SpiderFormatedMessage: &s}, nil
}
```

#### Step 4 编写拆句方法

``` go
import "proto" // 导入共享的数据结构契约包
// import ...

// 作为业务逻辑开发时，用以下代码注册
processor.Register("SplitArticle", func(config string) (processor.Processor, error) {
        return SplitArticle, nil
})

/* 作为插件时，用以下代码注册
var Bundle = plugin.Bundle(
	processor.Plugin("SplitArticle", func(config string) (processor.Processor, error) {
        return SplitArticle, nil
    }),
)
*/

type Request struct{
    // 注入NewtonArticleMessage
    NewtonArticleMessage *NewtonArticleMessage `inject:"NewtonArticleMessage"`
}

type Response struct{
    // 返回Sentences注入容器中
    NewtonSentenceObjects []*NewtonSetenceObject `inject:"NewtonSentenceObjects"`
}

func (h *Processor) SplitArticle(r *Request) *Response {
    ss := strings.Split(r.Content, ".")

    var sos []*proto.NewtonSetenceObject
    for _, s := range ss {
        sos = append(sos, &proto.NewtonSentenceObject{
            Sentence: s,
            BundleKey: r.BundleKey,
            PubDate: r.PubDate,
        })
    }

    return &Response{NewtonSentenceObjects: sos}
}

```

#### Step 5 编写将数据写入ES的方法


``` go
// 作为业务逻辑开发时，用以下代码注册
processor.Register("WriteSentences2ES", ProcessorFactory)

/* 作为插件时，用以下代码注册
var Bundle = plugin.Bundle(
	processor.Plugin("WriteSentences2ES", ProcessorFactory),
)
*/

type Request struct{
    // 注入上下文context.Context对象
    Ctx                 context.Context        `inject:"Ctx"`
    // 注入[]*NewtonSetenceObject
    Sentences           []*NewtonSetenceObject `inject:"NewtonSentenceObjects"`
    // 注入ESNewtonIssueClient
    ESNewtonDailyClient *elastic.Client        `inject:"ESNewtonDailyClient"`
    // 注入由解码Processor产生的NewtonArticleMessage
    ESNewtonIssueClient *elastic.Client        `inject:"ESNewtonIssueClient"`
}

type Config struct{
    NewtonDailyIndex string `json:"NewtonDailyIndex"` // 配置正常数据写入的ES index名字
    NewtonIssueIndex string `json:"NewtonIssueIndex"` // 配置异常数据写入的ES index名字
}

type Processor struct{
    config Config
}

func (h *Processor) WriteSentences2ES(r *Request) error {
    var err error
    for i, _ := range r.Sentences{
        sentence := r.Sentences[i]
        if time.Parse("2006-01-02 15:04:05", sentence.PubDate).After(time.Now()) {
            err = r.ESNewtonDailyClient.Index().
                Index(h.config.NewtonDailyIndex).
                BodyJson(sentence).
                Do(r.Ctx)
        } else {
            err = r.ESNewtonIssueClient.Index().
                Index(h.config.NewtonIssueIndex).
                BodyJson(sentence).
                Do(r.Ctx)
        }
    }
    return err
}

func ProcessorFactory(rawConfig string) (processor.Processor, error) {
    var conf Config
    err := json.Unmarshal([]byte(rawConfig), &conf)
    if err != nil {
        return nil, err
    }
    // 返回Processor方法
    return &Processor{config:conf}.WriteSentences2ES, nil
}
```

#### Step 6 编写将数据写入Kafka的方法

``` go

// 作为业务逻辑开发时，用以下代码注册
processor.Register("WriteSentences2Kafka", ProcessorFactory)

/* 作为插件时，用以下代码注册
var Bundle = plugin.Bundle(
	processor.Plugin("WriteSentences2Kafka", ProcessorFactory),
)
*/

type Request struct{
    // 注入kafka生成者实例
    Producer  sarama.SyncProducer    `inject:"KafkaNewtonSentenceProducer"`
    // 注入[]*NewtonSetenceObject
    Sentences []*NewtonSetenceObject `inject:"NewtonSentenceObjects"`
}

type Config struct{
    Topic string `json:"Topic"`
    Key   string `json:"Key"`
}

func Processor struct{
    config Config
}

func (h *Processor) WriteSentences2Kafka(r *Request) error {
    b, err := json.Marshal(r.NewtonArticleMessage)
    if err != nil {
        return error
    }

    var msgs []*sarama.ProducerMessage
    for i, _ := range r.Sentences {
        msgs = append(msgs, &sarama.ProducerMessage{
            Topic: h.config.Topic,
            Key: h.config.Key,
            Val: sarama.ByteEncoder(data),
        })
    }

    return h.Producer.SendMessages(msgs)
}

func ProcessorFactory(rawConfig string) (processor.Processor, error) {
    var conf Config
    err := json.Unmarshal([]byte(rawConfig), &conf)
    if err != nil {
        return nil, err
    }
    // 返回Processor方法
    return &Processor{config:conf}.WriteSentences2Kafka, nil
}
```


### Processor参数限制的原因

Processor定义：
``` go
    type Processor interface{}

    func Validate(processor Processor) error {
	    if reflect.TypeOf(processor).Kind() != reflect.Func {
		    return errors.New("Processor must be a callable func")
	    }
	    return nil
    }
```

例子: 它可以表现成一下任意情况
``` go
    func Handle(...Struct) (...Struct, error) {}

    // 例如
    func Handle() {}
    func Handle(Request) Response {}
    func Handle(*Request) (*Response, error) {}
    ...
```


设计原因: golang的反射只无法获得参数的名字，它不关心参数名字，只关心参数的类型和参数的顺序
而在设计时，我们可能会有多个同类型的不同作用的多个组件。

例如：我们从kafka读取数据，双写至不同的es，它们的Go Type是一致的，而仅仅是实例不一致
在基于https://github.com/codegangsta/inject 作为注入容器时，它以Go Type为key，Go Value作为value
存至map时，会发生碰撞，而导致同类型的值被覆盖。所以在此基础上我将其结构修改为
``` go
    map[reflect.Type]reflect.Value => map[reflect.Type]map[string]reflect.Value

    // 例如有如下结构体:
    // 1. 不带有inject的tag字段不会进行注入
    // 2. inject的tag没有值的情况下，会以StructField的Name尝试注入
    // 3. inject的tag有值的情况下，会以tag查找并注入
    type TestStruct struct {
        Ctx      context.Context     `inject`
        ESClient *elastic.Client     `inject`
        DB1      *sql.DB             `inject:"UserDB"`
        DB2      *sql.DB             `inject:"GoodsDB"`
        Consumer *cluster.Consumer   `inject:"KafkaOrderSyncConsumer"`
        Producer sarama.SyncProducer `inject:"KafkaSMSSyncProducer"`
        Topic    string              `inject`
        Offsets  int                 `inject`
        Noop     *sql.DB
    }

    /* 容器中实际保存的结构示例:
    map[reflect.Type]map[string]reflect.Value{
        reflect.TypeOf((*context.Context)(nil)) : map[string]reflect.Value{
            "Ctx": reflect.ValueOf(TestStruct.Ctx),
        },
        reflect.TypeOf(*elastic.Client) : map[string]reflect.Value{
            "ESclient" : reflect.ValueOf(TestStruct.ESClient),
        },
        reflect.TypeOf(*sql.DB) : map[string]reflect.Value{
            "UserDB" : reflect.ValueOf(TestStruct.DB1),
            "GoodsDB" : reflect.ValueOf(TestStruct.DB2),
        },
        reflect.TypeOf(*cluster.Consumer) : map[string]reflect.Value{
            "KafkaOrderSyncConsumer" : reflect.ValueOf(TestStruct.Consumer),
        },
        reflect.TypeOf((*sarama.SyncProducer)(nil)) : map[string]reflect.Value{
            "KafkaSMSSyncProducer" : reflect.ValueOf(TestStruct.Producer),
        },
        reflect.TypeOf(string) : map[string]reflect.Value{
            "Topic" : reflect.ValueOf(TestStruct.Topic),
        },
        reflect.TypeOf(int) : map[string]reflect.Value{
            "Offsets" : reflect.ValueOf(TestStruct.Offsets),
        },
    }
    */
```
在类型相同的基础上，在辅以实例的名称来做区分，满足同类型不同实例的注入
https://golang.org/ref/spec#Function_types
https://stackoverflow.com/questions/31377433/getting-method-parameter-names
