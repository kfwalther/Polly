import React, { useState, useEffect } from 'react';
import StockList from './components/StockList'
import { Header } from './components/Header';
//import { Helmet } from 'react-helmet'
import './App.css'

const LOCAL_STORAGE_KEY = 'L0C@L'

function App() {
  const [stocks, updateStock] = useState([])
  const [curSortBy, updateSortBy] = useState('LargestHoldings')
  // Define the sort-by options.
  const sortByOptions = [
    { label: 'Largest Holdings', value: 'LargestHoldings', col: 'marketValue', ascending: false },
    { label: 'Smallest Holdings', value: 'SmallestHoldings', col: 'marketValue', ascending: true },
    { label: 'Biggest Absolute Gainers', value: 'BiggestAbsGainers', col: 'unrealizedGains', ascending: false },
    { label: 'Biggest Absolute Losers', value: 'BiggestAbsLosers', col: 'marketValue', ascending: true },
  ];

  // Pull the sort-by settings when page reloads.
  useEffect(() => {
    const prevSortBy = JSON.parse(
      localStorage.getItem(LOCAL_STORAGE_KEY) ?? "[]"
    );
    if (prevSortBy) updateSortBy(prevSortBy)
    console.log('Loading sort by...' + curSortBy)
    // Look up the full sortBy object in the list by value (stored in useState).
    const sortByObj = sortByOptions.find(x => x.value === curSortBy);
    console.log('Sort column: ' + sortByObj.col)
    // Apply the sorting, use global 'window' variable to reference the StockList object.
    window.stockList.applySorting(sortByObj);

  }, [curSortBy])

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
      {/* <Helmet>
        <style>{'body { background-color: black; }'}</style>
      </Helmet> */}
      <Header />
      <label>
        <select value={curSortBy} onChange={handleSortByChange}>
          {sortByOptions.map((option) => (
            <option key={option.value} value={option.value}>{option.label}</option>
          ))}
        </select>
      </label>
      <p style={{ color: 'green' }}>Sorting by {curSortBy}</p>
      <button onClick={refreshMarketData}>Refresh</button>
      <StockList stockList={stocks} />
    </>
  )
}

export default App;
