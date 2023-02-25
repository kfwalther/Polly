import React, { useState } from 'react';
import MainPage from './components/MainPage'
import { Header } from './components/Header';
import { Helmet } from 'react-helmet'
import './App.css'

function App() {
  // This is what gets rendered on the page.
  return (
    <>
      <Helmet>
        <style>{'body { background-color: black; }'}</style>
      </Helmet>
      <Header />
      <MainPage />
    </>
  )
}

export default App;
