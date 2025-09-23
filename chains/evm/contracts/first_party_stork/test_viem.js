import { createPublicClient, http, getContract } from 'viem';
import { localhost } from 'viem/chains';
import fs from 'fs';

// TODO: temp test file for getters

// Create client
const client = createPublicClient({
  chain: localhost,
  transport: http('http://localhost:8545')
});

// Read the ABI from artifacts
const artifactPath = './artifacts/contracts/UpgradeableFirstPartyStork.sol/UpgradeableFirstPartyStork.json';
const artifact = JSON.parse(fs.readFileSync(artifactPath, 'utf8'));

// Contract setup
const contractAddress = '0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512';
const publisherAddress = '0x99e295e85cb07C16B7BB62A44dF532A7F2620237';

const contract = getContract({
  address: contractAddress,
  abi: artifact.abi,
  client
});

async function testGetters() {
  console.log('Testing FirstParty Stork Contract Getters');
  console.log('==========================================');

  try {
    // Test 1: Get registered publisher
    console.log('\n1. Testing getPublisherUser...');
    const publisher = await contract.read.getPublisherUser([publisherAddress]);
    console.log('✅ Publisher found:', publisher);
  } catch (error) {
    console.log('❌ Publisher not found:', error.message);
  }

  try {
    // Test 2: Get current round ID
    console.log('\n2. Testing getCurrentRoundId...');
    const roundId = await contract.read.getCurrentRoundId([publisherAddress, 'MY_RANDOM_VALUE']);
    console.log('✅ Current round ID:', roundId.toString());
  } catch (error) {
    console.log('❌ Error getting round ID:', error.message);
  }

  try {
    // Test 3: Get historical count
    console.log('\n3. Testing getHistoricalRecordsCount...');
    const count = await contract.read.getHistoricalRecordsCount([publisherAddress, 'MY_RANDOM_VALUE']);
    console.log('✅ Historical records count:', count.toString());
  } catch (error) {
    console.log('❌ Error getting historical count:', error.message);
  }

  try {
    // Test 4: Get latest value
    console.log('\n4. Testing getLatestTemporalNumericValue...');
    const value = await contract.read.getLatestTemporalNumericValue([publisherAddress, 'MY_RANDOM_VALUE']);
    console.log('✅ Latest value:', value);
  } catch (error) {
    console.log('❌ No latest value (expected):', error.message);
  }

  try {
    // Test 5: Get historical value
    console.log('\n5. Testing getHistoricalTemporalNumericValue...');
    const value = await contract.read.getHistoricalTemporalNumericValue([publisherAddress, 'MY_RANDOM_VALUE', 2]);
    console.log('✅ Historical value:', value);
  } catch (error) {
    console.log('❌ No historical value (expected):', error.message);
  }

//   try {
//     // Test 6: Try non-existent publisher
//     console.log('\n5. Testing with non-existent publisher...');
//     await contract.read.getPublisherUser(['0x0000000000000000000000000000000000000001']);
//   } catch (error) {
//     console.log('✅ Correctly reverted for non-existent publisher');
//   }

  console.log('\n==========================================');
  console.log('Getter tests completed!');
}

testGetters().catch(console.error);
