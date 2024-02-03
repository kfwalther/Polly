import { Container, Nav, Navbar } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import './NavHeader.css'

// Returns the navigation bar header.
export const NavHeader = () => {

  return (
    <Navbar>
      <Container>
        <Nav className="nav-header">
          {/* Use Link from react-router-dom to define the routing links. */}
          <Link to="/home/stock" className="nav-item">Stocks</Link>
          <Link to="/home/etf" className="nav-item">ETFs</Link>
          <Link to="/home/crypto" className="nav-item">Crypto</Link>
          <Link to="/home/full" className="nav-item">Full Portfolio</Link>
          <Link to="/transactions" className="nav-item">Transactions & Performance</Link>
          <Link to="/refresh" className="nav-item">Refresh Data</Link>
        </Nav>
      </Container>
    </Navbar>
  );
}
