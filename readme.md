
===
![设计要求](设计要求.png)

===

	编译器:go1.7+
	$exam:go build
	
	运行：
	
	单机模式：
		$./exam -url=http://pythoninterview.ulaiber.com/
		
	多机联合 （以host1,host2,host3为例）
		host1: 
			$./exam -port=8888
		host2:
			$./exam -port=8888
		host3:
			$./exam -url=http://pythoninterview.ulaiber.com/ -addrs="$host1:8888,$host2:8888" -port=8888
			
		执行的最终结果在host3显示

=====

