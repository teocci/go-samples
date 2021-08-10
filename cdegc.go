package main

import "fmt"

func rawToCentiCelsius(rawIn int16) int16 {
	const offset int16 = 125
	cDegC := [...]int16{7782, 7106, 6573, 6136, 5765, 5446, 5165, 4915, 4690, 4486, 4127, 3820, 3431, 3207, 2911, 2654, 2500, 2427, 2289, 2160, 2040, 1927, 1820, 1623, 1404, 1032, 725, 465, 239, 40, -297, -576, -813, -1018, -1199, -1874}
	raw := [...]int16{321, 393, 463, 530, 595, 658, 718, 776, 832, 887, 990, 1087, 1222, 1305, 1421, 1528, 1595, 1627, 1689, 1747, 1803, 1857, 1907, 2002, 2110, 2296, 2450, 2579, 2690, 2785, 2942, 3065, 3165, 3247, 3315, 3540}

	rawIn = rawIn + offset

	if rawIn < raw[1] {
		return 8000 // out of range -> very hot
	}

	for index, element := range raw {
		if rawIn < raw[index] {
			lower := raw[index-1]
			upper := element
			spread := upper - lower
			lowerWeight := spread - (rawIn - lower)
			upperWeight := spread - (upper - rawIn)

			weightedAvg := (lowerWeight*cDegC[index-1] + upperWeight*cDegC[index]) / (spread)

			return weightedAvg
		}
	}

	return -2000 // out of range -> very cold
}

func main() {
	fmt.Println(rawToCentiCelsius(4404))
}
