import './App.css'
import React from 'react';
import MainPage from './components/MainPage'
import { Header } from './components/Header';
import { NavHeader } from './components/NavHeader';
import { Routes, Route } from 'react-router-dom';
import TransactionsPage from './components/TransactionsPage';

function App() {
  // This is what gets rendered on the page.
  return (
    <>
      { /* Display the decorative header, and navigation bar. */}
      <Header />
      <NavHeader />
      { /* Based on the selected route path, load a specific page. Index page is the default. */}
      <Routes>
        <Route path="/*">
          <Route index element={<MainPage />} />
          <Route path='transactions' element={<TransactionsPage />} />
        </Route>
      </Routes>
    </>
  )
}

export default App;
