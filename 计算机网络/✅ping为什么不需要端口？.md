# ✅ping为什么不需要端口？

# 典型回答


[✅ping的原理是什么？](https://www.yuque.com/hollis666/ukxanr/ivry7a)



ping 命令在网络诊断中是一种常用工具，用于测试目标主机的连通性和响应时间。



ping 命令是基于 ICMP 协议的，ICMP 是 IP 协议的一部分，用于传递控制消息。ping命令本身处于应用层，相当于一个应用程序。它使用的ICMP协议是一个网络层协议。



但是，我们通常用到的端口号，其实是传输层（如 TCP 和 UDP）的一部分，用于区分同一 IP 地址上的多个服务或应用程序。



**由于，ping是一个应用层直接使用网络层协议的例子，不涉及到传输层，所以不需要指定端口号。**



> 更新: 2024-09-13 21:32:23  
> 原文: <https://www.yuque.com/hollis666/ukxanr/pfmnefsmxrwhzd81>