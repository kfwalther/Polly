import React, { useState, useEffect } from 'react';
import StockList from './StockList'
import { Helmet } from 'react-helmet'

const LOCAL_STORAGE_KEY = 'L0C@L'

function App() {
  const [stocks, updateStock] = useState([])
  const [curSortBy, updateSortBy] = useState('')
  // Define the sort-by options.
  const sortByOptions = [
    { label: 'Largest Holdings', value: 'LargestHoldings' },
    { label: 'Smallest Holdings', value: 'SmallestHoldings' },
    { label: 'Biggest Absolute Gainers', value: 'BiggestAbsGainers' },
    { label: 'Biggest Absolute Losers', value: 'BiggestAbsLosers' },
  ];

  // Pull the sort-by settings when page reloads.
  useEffect(() => {
    const prevSortBy = JSON.parse(
      localStorage.getItem(LOCAL_STORAGE_KEY) ?? "[]"
    );
    if (prevSortBy) updateSortBy(prevSortBy)
    console.log('Loading sort by...' + curSortBy)
    // TODO Apply sort setting
  }, [])

  // Use this method to refresh the market data.
  function refreshMarketData(e) {
    console.log('Refreshing data...')
    // TODO Call method to refresh our market data.
  }

  function handleSortByChange(e) {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(e.target.value))
    updateSortBy(e.target.value)
  }

  // This is what gets rendered on the page.
  return (
    <>
      <Helmet>
        <title>Polly - Portfolio Viewer</title>
        <style>{'body { background-color: black; }'}</style>
      </Helmet>
      <label>
        <select value={curSortBy} onChange={handleSortByChange}>
          {sortByOptions.map((option) => (
            <option key={option.value} value={option.value}>{option.label}</option>
          ))}
        </select>
      </label>
      <p style={{ color: 'green' }}>Sorting by {curSortBy}</p>
      <StockList stockList={stocks} />
      <button onClick={refreshMarketData}>Refresh</button>
    </>
  )
}

export default App;
