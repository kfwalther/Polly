import Container from 'react-bootstrap/Container';
import Nav from 'react-bootstrap/Nav';
import Navbar from 'react-bootstrap/Navbar';
import { Link } from 'react-router-dom';

export const NavHeader = () => {
    return (
        <Container>
            <Navbar expand="lg" bg="primary" variant="light">
                <Nav className="nav-header">
                    <Link to="/" className="nav-item" href="#home">Home</Link>
                    <Link to="/transactions" className="nav-item" href="#txns">Transactions</Link>
                </Nav>
            </Navbar>
        </Container>
    );
}
