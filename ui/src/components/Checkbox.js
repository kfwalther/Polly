// Define a custom Checkbox object to accept and label and onClick callback.
const Checkbox = ({ label, checked, onClick, ...props }) => {
    const defaultChecked = checked ? checked : false;

    return (
        <div className="checkbox-wrapper">
            <label>
                <input
                    type="checkbox"
                    checked={defaultChecked}
                    onChange={() => onClick(!checked)}
                />
                <span>{label}</span>
            </label>
        </div>
    );
};

export default Checkbox;