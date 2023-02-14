import React, { useState } from 'react';
import StockList from './components/StockList'
import { Header } from './components/Header';
import './App.css'

const LOCAL_STORAGE_KEY = 'L0C@L'

function App() {
  const [stocks] = useState([])

  // Use this method to refresh the market data.
  function refreshMarketData(e) {
    console.log('Refreshing data...')
    // TODO Call method to refresh our market data.
  }

  // This is what gets rendered on the page.
  return (
    <>
      {/* <Helmet>
        <style>{'body { background-color: black; }'}</style>
      </Helmet> */}
      <Header />
      <button onClick={refreshMarketData}>Refresh</button>
      <StockList stockList={stocks} />
    </>
  )
}

export default App;
