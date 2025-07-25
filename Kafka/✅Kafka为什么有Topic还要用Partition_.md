# ✅Kafka 为什么有 Topic 还要用 Partition?

# 典型回答


Topic和Partition是kafka中比较重要的概念。



> 主题：Topic是Kafka中承载消息的逻辑容器。可以理解为一个消息队列。生产者将消息发送到特定的Topic，消费者从Topic中读取消息。Topic可以被认为是逻辑上的消息流。在实际使用中多用来区分具体的业务。
>
> 分区：Partition。是Topic的物理分区。一个Topic可以被分成多个Partition，每个Partition是一个有序且持久化存储的日志文件。每个Partition都存储了一部分消息，并且有一个唯一的标识符（称为Partition ID）。
>



看上去，这两个都是存储消息的载体，那为啥要分两层呢，有了Topic还需要Partition干什么呢？



在软件领域中，任何问题都可以加一个中间层来解决，而这，就是类似的思想，在Topic的基础上，再细粒度的划分出了一层，主要能带来以下几个好处：



1. 提升吞吐量：通过将一个Topic分成多个Partition，可以实现消息的并行处理。每个Partition可以由不同的消费者组进行独立消费，这样就可以提高整个系统的吞吐量。



2. 负载均衡：Partition的数量通常比消费者组的数量多，这样可以使每个消费者组中的消费者均匀地消费消息。当有新的消费者加入或离开消费者组时，可以通过重新分配Partition的方式进行负载均衡。



3. 扩展性：通过增加Partition的数量，可以实现Kafka集群的扩展性。更多的Partition可以提供更高的并发处理能力和更大的存储容量。



综上，Topic是逻辑上的消息分类，而Partition是物理上的消息分区。通过将Topic分成多个Partition，可以实现提升吞吐量、负载均衡、以及增加可扩展性。



> 更新: 2024-09-13 21:31:23  
> 原文: <https://www.yuque.com/hollis666/ukxanr/opxlb0a177ehqyty>