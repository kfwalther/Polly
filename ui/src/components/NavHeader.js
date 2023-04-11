import Container from 'react-bootstrap/Container';
import Nav from 'react-bootstrap/Nav';
import Navbar from 'react-bootstrap/Navbar';
import { Link } from 'react-router-dom';

// Returns the navigation bar header.
export const NavHeader = () => {
    return (
        <Container>
            {/* Use react-bootstrap for the navigation bar. */}
            <Navbar >
                <Nav className="nav-header">
                    {/* Use Link from react-router-dom to define the routing links. */}
                    <Link to="/" className="nav-item">Home</Link>
                    <Link to="/transactions" className="nav-item">Transactions</Link>
                    <Link to="/history" className="nav-item">Portfolio History</Link>
                </Nav>
            </Navbar>
        </Container>
    );
}
