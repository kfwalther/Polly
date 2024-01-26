import './App.css'
import { BrowserRouter } from 'react-router-dom';
import { Header } from './components/Header';
import { NavHeader } from './components/NavHeader';
import { Routes, Route } from 'react-router-dom';
import MainPage from './components/MainPage'
import RefreshPage from './components/RefreshPage'
import TransactionsPage from './components/TransactionsPage';

function App() {
  // This is what gets rendered on the page.
  return (
    <BrowserRouter>
      { /* Display the decorative header, and navigation bar. */}
      <Header />
      <NavHeader />
      { /* Based on the selected route path, load a specific page. Index page is the default. */}
      <Routes>
        <Route exact path='/' element={<MainPage/>} />
        <Route path='/home/:category' element={<MainPage/>} />
        <Route path='/transactions' element={<TransactionsPage />} />
        <Route path='/refresh' element={<RefreshPage/>} />
      </Routes>
    </BrowserRouter>
  )
}

export default App;
