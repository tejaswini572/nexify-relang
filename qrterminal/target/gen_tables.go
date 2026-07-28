package main

import (
	"fmt"
	"os"

	"rsc.io/qr/coding"
)

func main() {
	f, _ := os.Create("tables.rs")
	defer f.Close()

	fmt.Fprintln(f, `// Generated tables
pub struct ECBlock {
    pub num_blocks: usize,
    pub data_codewords: usize,
}

pub struct ECInfo {
    pub ec_codewords_per_block: usize,
    pub blocks1: ECBlock,
    pub blocks2: Option<ECBlock>,
}

pub struct VersionInfo {
    pub version: usize,
    pub total_data_bytes: [usize; 4],
    pub ec: [ECInfo; 4], // L, M, Q, H
}
`)

	fmt.Fprintln(f, `pub const VERSIONS: [VersionInfo; 41] = [`)
	
	// Print a dummy v0
	fmt.Fprintln(f, `    VersionInfo { version: 0, total_data_bytes: [0,0,0,0], ec: [
		ECInfo { ec_codewords_per_block: 0, blocks1: ECBlock { num_blocks: 0, data_codewords: 0 }, blocks2: None },
		ECInfo { ec_codewords_per_block: 0, blocks1: ECBlock { num_blocks: 0, data_codewords: 0 }, blocks2: None },
		ECInfo { ec_codewords_per_block: 0, blocks1: ECBlock { num_blocks: 0, data_codewords: 0 }, blocks2: None },
		ECInfo { ec_codewords_per_block: 0, blocks1: ECBlock { num_blocks: 0, data_codewords: 0 }, blocks2: None }
	]},`)

	for v := 1; v <= 40; v++ {
		version := coding.Version(v)
		fmt.Fprintf(f, "    VersionInfo { version: %d, total_data_bytes: [", v)
		for l := 0; l < 4; l++ {
			fmt.Fprintf(f, "%d,", version.DataBytes(coding.Level(l)))
		}
		fmt.Fprintln(f, "], ec: [")
		
		for l := 0; l < 4; l++ {
			// In rsc.io/qr/coding, we need the block specs. 
			// Wait, rsc.io/qr/coding hides the block breakdown in unexported fields!
			// How can I get the block specs? 
			// Actually, rsc.io/qr has a method: 
			// plan := coding.NewPlan(version, level, mode)
			// we can extract it from plan!
		}
	}
}
