package main

import (
	"log"
)

func combin(key string) []string {
	list := make([]string, len(key))
	for i := range key {
		list[i] = string(key[i])
	}

	out := outOrder(list)

	m := make(map[string]struct{}, len(out))

	result := make([]string, 0, len(out))

	for i := range out {
		_, exist := m[out[i]]
		if !exist {
			m[out[i]] = struct{}{}
			result = append(result, out[i])
		}
	}

	return result
}

//输入trainsNums，返回全部排列
//如输入[1 2 3]，则返回[123 132 213 231 312 321]
func outOrder(trainsNums []string) []string {
	n := len(trainsNums)
	//检查
	if n == 0 || n > 10 {
		log.Println("Illegal argument")
		return nil
	}
	//如果只有一个数，则直接返回
	if n == 1 {
		return []string{trainsNums[0]}
	}
	//否则，将最后一个数插入到前面的排列数中的所有位置（递归）
	return insert(outOrder(trainsNums[:n-1]), trainsNums[n-1])
}
func insert(res []string, insertNum string) []string {
	//保存结果的slice
	result := make([]string, len(res)*(len(res[0])+1))
	index := 0
	for _, v := range res {
		for i := 0; i < len(v); i++ {
			//在v的每一个元素前面插入
			result[index] = v[:i] + insertNum + v[i:]
			index++
		}
		//在v最后面插入
		result[index] = v + insertNum
		index++
	}
	return result
}
