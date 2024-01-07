import Container from 'react-bootstrap/Container';
import { Nav, Navbar, NavDropdown } from 'react-bootstrap';
import { Link, NavLink } from 'react-router-dom';
import './NavHeader.css'

// Returns the navigation bar header.
export const NavHeader = () => {
    return (
        <Container>
            {/* Use react-bootstrap for the navigation bar. */}
            <Navbar >
                <Navbar.Toggle aria-controls="basic-navbar-nav" />
                <Navbar.Collapse id="basic-navbar-nav">
                    <Nav className="nav-header">
                        {/* Use Link from react-router-dom to define the routing links. */}
                        <NavDropdown title="Home" id="basic-nav-dropdown">
                            <NavDropdown.Item as={NavLink} to="/home/stock">Stocks</NavDropdown.Item>
                            <NavDropdown.Item as={NavLink} to="/home/etf">ETFs</NavDropdown.Item>
                            <NavDropdown.Item as={NavLink} to="/home/full">Full</NavDropdown.Item>
                        </NavDropdown>
                        <Link to="/transactions" className="nav-item">Transactions & Performance</Link>
                        <Link to="/refresh" className="nav-item">Refresh Data</Link>
                    </Nav>
                </Navbar.Collapse>
            </Navbar>
        </Container>
    );
}
