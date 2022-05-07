package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	. "github.com/ayushmaheshwari768/spartan-go"
)

func main() {
	fmt.Println("Starting simulation.  This may take a moment...")

	fakeNet := NewFakeNet(&FakeNet{})

	alice := NewClient(&Client{Name: "Alice", Net: fakeNet})
	bob := NewClient(&Client{Name: "Bob", Net: fakeNet})
	charlie := NewClient(&Client{Name: "Charlie", Net: fakeNet})

	minnie := NewMiner(&Client{Name: "Minnie", Net: fakeNet})
	mickey := NewMiner(&Client{Name: "Mickey", Net: fakeNet})

	genesis, err := MakeGenesis(&Blockchain{
		ClientBalanceMap: map[*Client]uint{
			alice:          uint(233),
			bob:            uint(99),
			charlie:        uint(67),
			&minnie.Client: uint(400),
			&mickey.Client: uint(300),
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	donald := NewMiner(&Client{Name: "Mickey", Net: fakeNet, StartingBlock: genesis}, 3000)

	showBalances := func(client *Client) {
		fmt.Println("Alice has " + strconv.FormatUint(uint64(client.LastBlock.BalanceOf(alice.Address)), 10) + " gold.")
		fmt.Println("Bob has " + strconv.FormatUint(uint64(client.LastBlock.BalanceOf(bob.Address)), 10) + " gold.")
		fmt.Println("Charlie has " + strconv.FormatUint(uint64(client.LastBlock.BalanceOf(charlie.Address)), 10) + " gold.")
		fmt.Println("Minnie has " + strconv.FormatUint(uint64(client.LastBlock.BalanceOf(minnie.Address)), 10) + " gold.")
		fmt.Println("Mickey has " + strconv.FormatUint(uint64(client.LastBlock.BalanceOf(mickey.Address)), 10) + " gold.")
		fmt.Println("Donald has " + strconv.FormatUint(uint64(client.LastBlock.BalanceOf(donald.Address)), 10) + " gold.")
	}

	fmt.Println("Initial balances:")
	showBalances(alice)

	fakeNet.Register(alice, bob, charlie, &minnie.Client, &mickey.Client, &donald.Client)

	minnie.Initialize()
	mickey.Initialize()

	fmt.Println("Alice is transferring 40 gold to " + bob.Address)
	alice.PostTransaction([]TxOuput{{Amount: 40, Address: bob.Address}})

	time.AfterFunc(time.Second*time.Duration(2000), func() {
		fmt.Println()
		fmt.Println("***Starting a late-to-the-party miner***")
		fmt.Println()
		fakeNet.Register(&donald.Client)
		donald.Initialize()
	})

	time.AfterFunc(time.Second*time.Duration(5000), func() {
		fmt.Println()
		fmt.Println("Minnie has a chain of length " + strconv.FormatUint(uint64(minnie.StartingBlock.ChainLength), 10))

		fmt.Println()
		fmt.Println("Mickey has a chain of length " + strconv.FormatUint(uint64(mickey.StartingBlock.ChainLength), 10))

		fmt.Println()
		fmt.Println("Donald has a chain of length " + strconv.FormatUint(uint64(donald.StartingBlock.ChainLength), 10))

		fmt.Println()
		fmt.Println("Final balances (Minnie's perspective):")
		showBalances(&minnie.Client)

		fmt.Println()
		fmt.Println("Final balances (Alice's perspective):")
		showBalances(alice)

		fmt.Println()
		fmt.Println("Final balances (Donald's perspective):")
		showBalances(&donald.Client)

		os.Exit(0)
	})
}
