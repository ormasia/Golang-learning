# ✅Kafka、ActiveMQ、RabbitMQ和RocketMQ都有哪些区别？

# 典型回答


Kafka、ActiveMQ、RabbitMQ和RocketMQ都是常见的消息中间件，它们都提供了高性能、高可用、可扩展的消息传递机制，但它们之间也有以下一些区别：



1. **消息传递模型**：Kafka主要支持发布-订阅模型，ActiveMQ、RabbitMQ和RocketMQ则同时支持点对点和发布-订阅两种模型。
2. **性能和吞吐量**：Kafka在数据处理和数据分发方面表现出色，可以处理每秒数百万条消息，而ActiveMQ、RabbitMQ和RocketMQ的吞吐量相对较低。
3. **消息分区和负载均衡**：Kafka将消息划分为多个分区，并分布在多个服务器上，实现负载均衡和高可用性。ActiveMQ、RabbitMQ和RocketMQ也支持消息分区和负载均衡，但实现方式不同，例如RabbitMQ使用了一种叫做Sharding的机制。
4. **开发和部署复杂度**：Kafka相对比较简单，易于使用和部署，但在实现一些高级功能时需要进行一些复杂的配置。ActiveMQ、RabbitMQ和RocketMQ则提供了更多的功能和选项，也更加灵活，但相应地会增加开发和部署的复杂度。
5. **社区和生态**：Kafka、ActiveMQ、RabbitMQ和RocketMQ都拥有庞大的社区和完善的生态系统，但Kafka和RocketMQ目前的发展势头比较迅猛，社区活跃度也相对较高。
6. **功能支持：**

****

| | #### 优先级队列 | **延迟队列** | **死信队列** | #### 重试队列 | **消费模式** | **事务消息** |
| --- | --- | --- | --- | --- | --- | --- |
| **Kafka** | 不支持 | <font style="color:rgb(55, 65, 81);">不支持，可以间接实现延迟队列</font> | 无 | <font style="color:rgb(55, 65, 81);">不直接支持，可以通过消费者逻辑来实现重试机制。</font> | <font style="color:rgb(55, 65, 81);">主要是拉模式。</font> | <font style="color:rgb(55, 65, 81);">支持事务，但限于消息生产。</font> |
| **RocketMQ** | 支持 | <font style="color:rgb(55, 65, 81);">直接支持延迟队列，可以设定消息的延迟时间。</font><br/>    | <font style="color:rgb(55, 65, 81);">支持</font> | <font style="color:rgb(55, 65, 81);">支持重试队列，可以自动或手动将消息重新发送。</font> | <font style="color:rgb(55, 65, 81);">支持推和拉两种模式。</font> | <font style="color:rgb(55, 65, 81);">支持事务消息。</font> |
| **RabbitMQ** | 支持 | <font style="color:rgb(55, 65, 81);">支持延迟队列，可以通过插件或者消息TTL和死信交换来实现。</font> | <font style="color:rgb(55, 65, 81);">支持</font> | <font style="color:rgb(55, 65, 81);">可以实现重试机制，但需要通过消息属性和额外配置来手动设置。</font> | <font style="color:rgb(55, 65, 81);">主要是推模式，但也可以实现拉模式。</font> | <font style="color:rgb(55, 65, 81);">支持基本的消息事务。</font> |
| **ActiveMQ** | 支持 | 支持 | <font style="color:rgb(55, 65, 81);">支持</font> | <font style="color:rgb(55, 65, 81);">支持重试机制，可以配置消息重发策略。</font> | <font style="color:rgb(55, 65, 81);">支持推和拉两种模式。</font> | <font style="color:rgb(55, 65, 81);">支持事务消息。</font> |




总的来说，这些消息中间件都有自己的优缺点，选择哪一种取决于具体的业务需求和系统架构。



# 扩展知识
## 如何选型


在选择消息队列技术时，需要根据实际业务需求和系统特点来选择，以下是一些参考因素：



1. 性能和吞吐量：如果需要处理海量数据，需要高性能和高吞吐量，那么Kafka是一个不错的选择。



2. 可靠性：如果需要保证消息传递的可靠性，包括数据不丢失和消息不重复投递，那么RocketMQ和RabbitMQ都提供了较好的可靠性保证。



3. 消息传递模型：如果需要支持发布-订阅和点对点模型，那么RocketMQ和RabbitMQ是一个不错的选择。如果只需要发布-订阅模型，Kafka则是一个更好的选择。



4. 消息持久化：如果需要更快地持久化消息，并且支持高效的消息查询，那么Kafka是一个不错的选择。如果需要更加传统的消息持久化方式，那么RocketMQ和RabbitMQ可以满足需求。



5. 开发和部署复杂度：Kafka比较简单，易于使用和部署，但在实现一些高级功能时需要进行一些复杂的配置。RocketMQ和RabbitMQ提供了更多的功能和选项，也更加灵活，但相应地会增加开发和部署的复杂度。



6. 社区和生态：Kafka、RocketMQ和RabbitMQ都拥有庞大的社区和完善的生态系统，但Kafka和RocketMQ目前的发展势头比较迅猛，社区活跃度也相对较高。



7. 实现语言方面，kafka是基于scala和java开发的，rocketmq、activemq等都是基于java语言的，rabbitmq是基于erlang的。



8. 功能性，上面列举过一些功能，我们在选型的时候需要看哪个可以满足我们的需求。



需要根据具体情况来选择最适合的消息队列技术。如果有多个因素需要考虑，可以进行性能测试和功能评估来辅助选择。  




> 更新: 2024-09-13 21:31:27  
> 原文: <https://www.yuque.com/hollis666/ukxanr/vst81qlgvl7yelgo>