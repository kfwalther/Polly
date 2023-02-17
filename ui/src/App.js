import React, { useState } from 'react';
import MainPage from './components/MainPage'
import { Header } from './components/Header';
import './App.css'

function App() {
  // This is what gets rendered on the page.
  return (
    <>
      <Header />
      <MainPage />
    </>
  )
}

export default App;
