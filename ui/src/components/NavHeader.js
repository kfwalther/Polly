import { Container, Nav, Navbar, NavDropdown } from 'react-bootstrap';
import { Link, NavLink } from 'react-router-dom';
import './NavHeader.css'

// Returns the navigation bar header.
export const NavHeader = () => {

  return (
    <Navbar>
      <Container>
        <Navbar.Toggle aria-controls="responsive-navbar-nav" />
        <Navbar.Collapse id="responsive-navbar-nav">
          <Nav className="nav-header">
            {/* Use Link from react-router-dom to define the routing links. */}
            <NavDropdown
              title="Home"
              id="collapsible-nav-dropdown"
              className="vertical-dropdown"
            >
              <NavDropdown.Item as={NavLink} to="/home/stock">Stocks</NavDropdown.Item>
              <NavDropdown.Item as={NavLink} to="/home/etf">ETFs</NavDropdown.Item>
              <NavDropdown.Item as={NavLink} to="/home/crypto">Crypto</NavDropdown.Item>
              <NavDropdown.Item as={NavLink} to="/home/full">Full</NavDropdown.Item>
            </NavDropdown>
            <Link to="/transactions" className="nav-item">Transactions & Performance</Link>
            <Link to="/refresh" className="nav-item">Refresh Data</Link>
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
}
