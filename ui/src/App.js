import React, { useState } from 'react';
import MainPage from './components/MainPage'
import { Header } from './components/Header';
import { NavHeader } from './components/NavHeader';
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
      <NavHeader />
      <MainPage />
    </>
  )
}

export default App;
