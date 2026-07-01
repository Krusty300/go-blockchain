package main

import "crypto/sha256"

type MerkleTree struct {
    RootNode *MerkleNode
}

type MerkleNode struct {
    Left  *MerkleNode
    Right *MerkleNode
    Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
    node := MerkleNode{}
    
    if left == nil && right == nil {
        hash := sha256.Sum256(data)
        node.Data = hash[:]
    } else {
        prevHashes := append(left.Data, right.Data...)
        hash := sha256.Sum256(prevHashes)
        node.Data = hash[:]
    }
    
    node.Left = left
    node.Right = right
    
    return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
    var nodes []MerkleNode
    
    // Create leaf nodes
    for _, datum := range data {
        node := NewMerkleNode(nil, nil, datum)
        nodes = append(nodes, *node)
    }
    
    // Handle odd number of transactions
    if len(nodes)%2 != 0 {
        nodes = append(nodes, nodes[len(nodes)-1])
    }
    
    // Build tree
    for len(nodes) > 1 {
        var level []MerkleNode
        
        for i := 0; i < len(nodes); i += 2 {
            node := NewMerkleNode(&nodes[i], &nodes[i+1], nil)
            level = append(level, *node)
        }
        
        nodes = level
    }
    
    return &MerkleTree{&nodes[0]}
}