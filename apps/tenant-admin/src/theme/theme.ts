import { createTheme } from '@mui/material';

export const theme = createTheme({
    palette: {
        primary: {
            main: '#1a73e8', // Google Blue
            light: '#e8f0fe',
            dark: '#174ea6',
        },
        success: {
            main: '#1e8e3e', // Google Green
        },
        warning: {
            main: '#f9ab00', // Google Yellow
        },
        background: {
            default: '#f8f9fa',
            paper: '#ffffff',
        },
        text: {
            primary: '#202124',
            secondary: '#5f6368',
        },
    },
    typography: {
        fontFamily: '"Google Sans", "Roboto", "Helvetica", "Arial", sans-serif',
        h5: {
            fontWeight: 500,
            letterSpacing: -0.5,
        },
        h6: {
            fontWeight: 500,
            fontSize: '1.1rem',
        },
        button: {
            textTransform: 'none',
            fontWeight: 500,
            fontSize: '0.875rem',
        },
    },
    shape: {
        borderRadius: 8,
    },
    components: {
        MuiButton: {
            styleOverrides: {
                root: {
                    padding: '8px 24px',
                    boxShadow: 'none',
                    '&:hover': {
                        boxShadow: '0 1px 2px 0 rgba(60,64,67,.302), 0 1px 3px 1px rgba(60,64,67,.149)',
                    },
                },
                containedPrimary: {
                    backgroundColor: '#1a73e8',
                }
            },
        },
        MuiCard: {
            styleOverrides: {
                root: {
                    border: '1px solid #dadce0',
                    boxShadow: 'none',
                    '&:hover': {
                        boxShadow: '0 1px 2px 0 rgba(60,64,67,.3), 0 1px 3px 1px rgba(60,64,67,.15)',
                    },
                },
            },
        },
    },
});
