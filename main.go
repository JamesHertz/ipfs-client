package main



func main(){

	ipfs := NewClient()
	ipfs.BootstrapNode()

	// let's print intesting metadata

	/*
//	info, err := sh.ID()

	if err != nil {
		panic(err)
	}

	fmt.Printf("MyNodeId          : %s\n", info.ID)
	fmt.Printf("MyProtocolVersion : %s\n", info.ProtocolVersion)

	for _, addr := range info.Addresses {
		if suitableMultiAddress(addr) {
			fmt.Println("->", addr)
		}
	}
*/

}