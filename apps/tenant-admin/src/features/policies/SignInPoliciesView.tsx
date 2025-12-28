import React, { useState } from 'react';
import {
    Box,
    Typography,
    Button,
    Card,
    CardContent,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    Grid,
    IconButton,
    Chip,
    FormControlLabel,
    Checkbox,
    MenuItem,
    Paper,
    CircularProgress,
    Divider,
    Stack,
    ToggleButton,
    ToggleButtonGroup,
    Tooltip,
    Autocomplete,
    createFilterOptions,
    TablePagination,
    useMediaQuery,
    useTheme
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
    CheckCircle as CheckCircleIcon,
    Block as BlockIcon,
    Warning as WarningIcon,
    DeviceHub as DeviceHubIcon,
    Public as PublicIcon,
    Group as GroupIcon,
    KeyboardArrowRight as ArrowIcon,
    PlayArrow as PlayArrowIcon,
    Pause as PauseIcon,
    Security as SecurityIcon
} from '@mui/icons-material';
import { SignInPolicy, PolicyNode, PolicyCondition } from '../../types';

interface SignInPoliciesViewProps {
    policies: SignInPolicy[];
    onRefresh: () => void;
}

const CONDITION_TYPES = [
    { value: 'User', label: 'User / Group', icon: <GroupIcon fontSize="small" /> },
    { value: 'Network', label: 'Network / IP', icon: <PublicIcon fontSize="small" /> },
    { value: 'Device', label: 'Device / OS', icon: <DeviceHubIcon fontSize="small" /> },
];

const OPS: Record<string, { label: string; value: string; field: string }[]> = {
    'User': [{ label: 'In Group', value: 'in_group', field: 'group' }, { label: 'Is User', value: 'is', field: 'email' }],
    'Network': [
        { label: 'IP in CIDR', value: 'cidr', field: 'ip' },
        { label: 'Country Equals', value: 'country', field: 'country' },
        { label: 'IP is Private', value: 'is_private', field: 'ip' }
    ],
    'Device': [{ label: 'OS Equals', value: 'os', field: 'os' }],
};

const OS_OPTIONS = [
    { value: 'windows', label: 'Windows (windows)' },
    { value: 'linux', label: 'Linux (linux)' },
    { value: 'darwin', label: 'macOS (darwin)' },
];

const COUNTRIES = [
    { code: 'AF', label: 'Afghanistan' },
    { code: 'AX', label: 'Aland Islands' },
    { code: 'AL', label: 'Albania' },
    { code: 'DZ', label: 'Algeria' },
    { code: 'AS', label: 'American Samoa' },
    { code: 'AD', label: 'Andorra' },
    { code: 'AO', label: 'Angola' },
    { code: 'AI', label: 'Anguilla' },
    { code: 'AQ', label: 'Antarctica' },
    { code: 'AG', label: 'Antigua and Barbuda' },
    { code: 'AR', label: 'Argentina' },
    { code: 'AM', label: 'Armenia' },
    { code: 'AW', label: 'Aruba' },
    { code: 'AU', label: 'Australia' },
    { code: 'AT', label: 'Austria' },
    { code: 'AZ', label: 'Azerbaijan' },
    { code: 'BS', label: 'Bahamas' },
    { code: 'BH', label: 'Bahrain' },
    { code: 'BD', label: 'Bangladesh' },
    { code: 'BB', label: 'Barbados' },
    { code: 'BY', label: 'Belarus' },
    { code: 'BE', label: 'Belgium' },
    { code: 'BZ', label: 'Belize' },
    { code: 'BJ', label: 'Benin' },
    { code: 'BM', label: 'Bermuda' },
    { code: 'BT', label: 'Bhutan' },
    { code: 'BO', label: 'Bolivia' },
    { code: 'BQ', label: 'Bonaire, Sint Eustatius and Saba' },
    { code: 'BA', label: 'Bosnia and Herzegovina' },
    { code: 'BW', label: 'Botswana' },
    { code: 'BV', label: 'Bouvet Island' },
    { code: 'BR', label: 'Brazil' },
    { code: 'IO', label: 'British Indian Ocean Territory' },
    { code: 'BN', label: 'Brunei Darussalam' },
    { code: 'BG', label: 'Bulgaria' },
    { code: 'BF', label: 'Burkina Faso' },
    { code: 'BI', label: 'Burundi' },
    { code: 'KH', label: 'Cambodia' },
    { code: 'CM', label: 'Cameroon' },
    { code: 'CA', label: 'Canada' },
    { code: 'CV', label: 'Cape Verde' },
    { code: 'KY', label: 'Cayman Islands' },
    { code: 'CF', label: 'Central African Republic' },
    { code: 'TD', label: 'Chad' },
    { code: 'CL', label: 'Chile' },
    { code: 'CN', label: 'China' },
    { code: 'CX', label: 'Christmas Island' },
    { code: 'CC', label: 'Cocos (Keeling) Islands' },
    { code: 'CO', label: 'Colombia' },
    { code: 'KM', label: 'Comoros' },
    { code: 'CG', label: 'Congo' },
    { code: 'CD', label: 'Congo, Democratic Republic of the' },
    { code: 'CK', label: 'Cook Islands' },
    { code: 'CR', label: 'Costa Rica' },
    { code: 'CI', label: 'Cote D\'Ivoire' },
    { code: 'HR', label: 'Croatia' },
    { code: 'CU', label: 'Cuba' },
    { code: 'CW', label: 'Curacao' },
    { code: 'CY', label: 'Cyprus' },
    { code: 'CZ', label: 'Czech Republic' },
    { code: 'DK', label: 'Denmark' },
    { code: 'DJ', label: 'Djibouti' },
    { code: 'DM', label: 'Dominica' },
    { code: 'DO', label: 'Dominican Republic' },
    { code: 'EC', label: 'Ecuador' },
    { code: 'EG', label: 'Egypt' },
    { code: 'SV', label: 'El Salvador' },
    { code: 'GQ', label: 'Equatorial Guinea' },
    { code: 'ER', label: 'Eritrea' },
    { code: 'EE', label: 'Estonia' },
    { code: 'ET', label: 'Ethiopia' },
    { code: 'FK', label: 'Falkland Islands' },
    { code: 'FO', label: 'Faroe Islands' },
    { code: 'FJ', label: 'Fiji' },
    { code: 'FI', label: 'Finland' },
    { code: 'FR', label: 'France' },
    { code: 'GF', label: 'French Guiana' },
    { code: 'PF', label: 'French Polynesia' },
    { code: 'TF', label: 'French Southern Territories' },
    { code: 'GA', label: 'Gabon' },
    { code: 'GM', label: 'Gambia' },
    { code: 'GE', label: 'Georgia' },
    { code: 'DE', label: 'Germany' },
    { code: 'GH', label: 'Ghana' },
    { code: 'GI', label: 'Gibraltar' },
    { code: 'GR', label: 'Greece' },
    { code: 'GL', label: 'Greenland' },
    { code: 'GD', label: 'Grenada' },
    { code: 'GP', label: 'Guadeloupe' },
    { code: 'GU', label: 'Guam' },
    { code: 'GT', label: 'Guatemala' },
    { code: 'GG', label: 'Guernsey' },
    { code: 'GN', label: 'Guinea' },
    { code: 'GW', label: 'Guinea-Bissau' },
    { code: 'GY', label: 'Guyana' },
    { code: 'HT', label: 'Haiti' },
    { code: 'HM', label: 'Heard Island and Mcdonald Islands' },
    { code: 'VA', label: 'Holy See (Vatican City State)' },
    { code: 'HN', label: 'Honduras' },
    { code: 'HK', label: 'Hong Kong' },
    { code: 'HU', label: 'Hungary' },
    { code: 'IS', label: 'Iceland' },
    { code: 'IN', label: 'India' },
    { code: 'ID', label: 'Indonesia' },
    { code: 'IR', label: 'Iran' },
    { code: 'IQ', label: 'Iraq' },
    { code: 'IE', label: 'Ireland' },
    { code: 'IM', label: 'Isle of Man' },
    { code: 'IL', label: 'Israel' },
    { code: 'IT', label: 'Italy' },
    { code: 'JM', label: 'Jamaica' },
    { code: 'JP', label: 'Japan' },
    { code: 'JE', label: 'Jersey' },
    { code: 'JO', label: 'Jordan' },
    { code: 'KZ', label: 'Kazakhstan' },
    { code: 'KE', label: 'Kenya' },
    { code: 'KI', label: 'Kiribati' },
    { code: 'KP', label: 'North Korea' },
    { code: 'KR', label: 'South Korea' },
    { code: 'KW', label: 'Kuwait' },
    { code: 'KG', label: 'Kyrgyzstan' },
    { code: 'LA', label: 'Laos' },
    { code: 'LV', label: 'Latvia' },
    { code: 'LB', label: 'Lebanon' },
    { code: 'LS', label: 'Lesotho' },
    { code: 'LR', label: 'Liberia' },
    { code: 'LY', label: 'Libya' },
    { code: 'LI', label: 'Liechtenstein' },
    { code: 'LT', label: 'Lithuania' },
    { code: 'LU', label: 'Luxembourg' },
    { code: 'MO', label: 'Macao' },
    { code: 'MK', label: 'Macedonia' },
    { code: 'MG', label: 'Madagascar' },
    { code: 'MW', label: 'Malawi' },
    { code: 'MY', label: 'Malaysia' },
    { code: 'MV', label: 'Maldives' },
    { code: 'ML', label: 'Mali' },
    { code: 'MT', label: 'Malta' },
    { code: 'MH', label: 'Marshall Islands' },
    { code: 'MQ', label: 'Martinique' },
    { code: 'MR', label: 'Mauritania' },
    { code: 'MU', label: 'Mauritius' },
    { code: 'YT', label: 'Mayotte' },
    { code: 'MX', label: 'Mexico' },
    { code: 'FM', label: 'Micronesia' },
    { code: 'MD', label: 'Moldova' },
    { code: 'MC', label: 'Monaco' },
    { code: 'MN', label: 'Mongolia' },
    { code: 'ME', label: 'Montenegro' },
    { code: 'MS', label: 'Montserrat' },
    { code: 'MA', label: 'Morocco' },
    { code: 'MZ', label: 'Mozambique' },
    { code: 'MM', label: 'Myanmar' },
    { code: 'NA', label: 'Namibia' },
    { code: 'NR', label: 'Nauru' },
    { code: 'NP', label: 'Nepal' },
    { code: 'NL', label: 'Netherlands' },
    { code: 'NC', label: 'New Caledonia' },
    { code: 'NZ', label: 'New Zealand' },
    { code: 'NI', label: 'Nicaragua' },
    { code: 'NE', label: 'Niger' },
    { code: 'NG', label: 'Nigeria' },
    { code: 'NU', label: 'Niue' },
    { code: 'NF', label: 'Norfolk Island' },
    { code: 'MP', label: 'Northern Mariana Islands' },
    { code: 'NO', label: 'Norway' },
    { code: 'OM', label: 'Oman' },
    { code: 'PK', label: 'Pakistan' },
    { code: 'PW', label: 'Palau' },
    { code: 'PS', label: 'Palestine' },
    { code: 'PA', label: 'Panama' },
    { code: 'PG', label: 'Papua New Guinea' },
    { code: 'PY', label: 'Paraguay' },
    { code: 'PE', label: 'Peru' },
    { code: 'PH', label: 'Philippines' },
    { code: 'PN', label: 'Pitcairn' },
    { code: 'PL', label: 'Poland' },
    { code: 'PT', label: 'Portugal' },
    { code: 'PR', label: 'Puerto Rico' },
    { code: 'QA', label: 'Qatar' },
    { code: 'RE', label: 'Reunion' },
    { code: 'RO', label: 'Romania' },
    { code: 'RU', label: 'Russia' },
    { code: 'RW', label: 'Rwanda' },
    { code: 'BL', label: 'Saint Barthelemy' },
    { code: 'SH', label: 'Saint Helena' },
    { code: 'KN', label: 'Saint Kitts and Nevis' },
    { code: 'LC', label: 'Saint Lucia' },
    { code: 'MF', label: 'Saint Martin' },
    { code: 'PM', label: 'Saint Pierre and Miquelon' },
    { code: 'VC', label: 'Saint Vincent and the Grenadines' },
    { code: 'WS', label: 'Samoa' },
    { code: 'SM', label: 'San Marino' },
    { code: 'ST', label: 'Sao Tome and Principe' },
    { code: 'SA', label: 'Saudi Arabia' },
    { code: 'SN', label: 'Senegal' },
    { code: 'RS', label: 'Serbia' },
    { code: 'SC', label: 'Seychelles' },
    { code: 'SL', label: 'Sierra Leone' },
    { code: 'SG', label: 'Singapore' },
    { code: 'SX', label: 'Sint Maarten' },
    { code: 'SK', label: 'Slovakia' },
    { code: 'SI', label: 'Slovenia' },
    { code: 'SB', label: 'Solomon Islands' },
    { code: 'SO', label: 'Somalia' },
    { code: 'ZA', label: 'South Africa' },
    { code: 'GS', label: 'South Georgia and the South Sandwich Islands' },
    { code: 'SS', label: 'South Sudan' },
    { code: 'ES', label: 'Spain' },
    { code: 'LK', label: 'Sri Lanka' },
    { code: 'SD', label: 'Sudan' },
    { code: 'SR', label: 'Suriname' },
    { code: 'SJ', label: 'Svalbard and Jan Mayen' },
    { code: 'SZ', label: 'Swaziland' },
    { code: 'SE', label: 'Sweden' },
    { code: 'CH', label: 'Switzerland' },
    { code: 'SY', label: 'Syria' },
    { code: 'TW', label: 'Taiwan' },
    { code: 'TJ', label: 'Tajikistan' },
    { code: 'TZ', label: 'Tanzania' },
    { code: 'TH', label: 'Thailand' },
    { code: 'TL', label: 'Timor-Leste' },
    { code: 'TG', label: 'Togo' },
    { code: 'TK', label: 'Tokelau' },
    { code: 'TO', label: 'Tonga' },
    { code: 'TT', label: 'Trinidad and Tobago' },
    { code: 'TN', label: 'Tunisia' },
    { code: 'TR', label: 'Turkey' },
    { code: 'TM', label: 'Turkmenistan' },
    { code: 'TC', label: 'Turks and Caicos Islands' },
    { code: 'TV', label: 'Tuvalu' },
    { code: 'UG', label: 'Uganda' },
    { code: 'UA', label: 'Ukraine' },
    { code: 'AE', label: 'United Arab Emirates' },
    { code: 'GB', label: 'United Kingdom' },
    { code: 'US', label: 'United States' },
    { code: 'UM', label: 'United States Minor Outlying Islands' },
    { code: 'UY', label: 'Uruguay' },
    { code: 'UZ', label: 'Uzbekistan' },
    { code: 'VU', label: 'Vanuatu' },
    { code: 'VE', label: 'Venezuela' },
    { code: 'VN', label: 'Vietnam' },
    { code: 'VG', label: 'Virgin Islands, British' },
    { code: 'VI', label: 'Virgin Islands, U.S.' },
    { code: 'WF', label: 'Wallis and Futuna' },
    { code: 'EH', label: 'Western Sahara' },
    { code: 'YE', label: 'Yemen' },
    { code: 'ZM', label: 'Zambia' },
    { code: 'ZW', label: 'Zimbabwe' },
].sort((a, b) => a.label.localeCompare(b.label));

const filter = createFilterOptions<{ label: string; value: string }>();

const NodeEditor: React.FC<{
    node: PolicyNode;
    stage: 'pre_auth' | 'post_auth';
    onChange: (newNode: PolicyNode) => void;
    onDelete?: () => void;
    depth?: number;
}> = ({ node, stage, onChange, onDelete, depth = 0 }) => {
    const isLeaf = !!node.condition;

    const handleAddCondition = () => {
        const defaultType = stage === 'pre_auth' ? 'Network' : 'User';
        const defaultOp = OPS[defaultType][0];
        const newNode: PolicyNode = {
            operator: 'AND',
            condition: { type: defaultType, field: defaultOp.field, op: defaultOp.value, value: '' }
        };
        if (node.children) {
            onChange({ ...node, children: [...node.children, newNode] });
        } else {
            // Transform to branch
            onChange({ ...node, condition: undefined, children: [newNode] });
        }
    };

    const handleAddSubGroup = () => {
        const newNode: PolicyNode = {
            operator: 'OR',
            children: []
        };
        if (node.children) {
            onChange({ ...node, children: [...node.children, newNode] });
        } else {
            onChange({ ...node, condition: undefined, children: [newNode] });
        }
    };

    const handleChildChange = (index: number, child: PolicyNode) => {
        const newChildren = [...(node.children || [])];
        newChildren[index] = child;
        onChange({ ...node, children: newChildren });
    };

    const handleRemoveChild = (index: number) => {
        const newChildren = [...(node.children || [])];
        newChildren.splice(index, 1);
        onChange({ ...node, children: newChildren });
    };

    const [identityOptions, setIdentityOptions] = useState<{ label: string; value: string; type: string }[]>([]);
    const [searching, setSearching] = useState(false);

    const handleSearchIdentity = async (query: string) => {
        if (!query || query.length < 2) return;
        setSearching(true);
        try {
            const res = await fetch(`/api/v1/identity/search?q=${encodeURIComponent(query)}`);
            const data = await res.json();
            if (data.success) {
                setIdentityOptions(data.data || []);
            }
        } catch (err) {
            console.error('Search failed', err);
        } finally {
            setSearching(false);
        }
    };

    return (
        <Paper variant="outlined" sx={{
            p: 2.5,
            mb: 2,
            border: '1px solid #e0e0e0',
            borderLeft: depth > 0 ? `6px solid ${node.operator === 'AND' ? '#1a73e8' : '#f9ab00'}` : '1px solid #e0e0e0',
            bgcolor: depth % 2 === 0 ? 'rgba(248, 249, 250, 0.3)' : '#fff',
            borderRadius: 3,
            boxShadow: depth === 0 ? '0 2px 4px rgba(0,0,0,0.02)' : 'none',
            '&:hover': {
                borderColor: depth > 0 ? (node.operator === 'AND' ? '#1a73e8' : '#f9ab00') : '#1a73e8',
            }
        }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: isLeaf ? 0 : 2, gap: 2 }}>
                {!isLeaf && (
                    <ToggleButtonGroup
                        size="small"
                        color="primary"
                        value={node.operator}
                        exclusive
                        onChange={(_, val) => val && onChange({ ...node, operator: val })}
                        sx={{
                            bgcolor: '#fff',
                            '& .MuiToggleButton-root': {
                                px: 2,
                                py: 0.5,
                                fontWeight: 800,
                                fontSize: '0.7rem',
                                borderRadius: 1.5,
                                border: '1px solid #f1f3f4',
                                '&.Mui-selected': {
                                    bgcolor: node.operator === 'AND' ? 'rgba(26, 115, 232, 0.1)' : 'rgba(249, 171, 0, 0.1)',
                                    color: node.operator === 'AND' ? '#1a73e8' : '#e37400',
                                    '&:hover': {
                                        bgcolor: node.operator === 'AND' ? 'rgba(26, 115, 232, 0.15)' : 'rgba(249, 171, 0, 0.15)',
                                    }
                                }
                            }
                        }}
                    >
                        <ToggleButton value="AND">AND</ToggleButton>
                        <ToggleButton value="OR">OR</ToggleButton>
                    </ToggleButtonGroup>
                )}

                {isLeaf && (
                    <Grid container spacing={1} alignItems="center" sx={{ flexGrow: 1, width: '100%' }}>
                        <Grid size={{ xs: 12, md: 2.5 }}>
                            <TextField
                                select
                                fullWidth
                                size="small"
                                label="Type"
                                value={node.condition?.type}
                                onChange={(e) => {
                                    const newType = e.target.value;
                                    const defaultOp = OPS[newType][0];
                                    onChange({ ...node, condition: { ...node.condition!, type: newType, op: defaultOp.value, field: defaultOp.field, value: '' } });
                                }}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            >
                                {CONDITION_TYPES.filter(t => stage === 'post_auth' ? (t.value === 'User' || t.value === 'Device') : t.value === 'Network').map(t => <MenuItem key={t.value} value={t.value} sx={{ gap: 1 }}>{t.icon} {t.label}</MenuItem>)}
                            </TextField>
                        </Grid>
                        <Grid size={{ xs: 12, md: 2.5 }}>
                            <TextField
                                select
                                fullWidth
                                size="small"
                                label="Operator"
                                value={node.condition?.op}
                                onChange={(e) => {
                                    const newOp = e.target.value;
                                    const opConfig = OPS[node.condition!.type].find(o => o.value === newOp);
                                    onChange({ ...node, condition: { ...node.condition!, op: newOp, field: opConfig?.field || '', value: '' } });
                                }}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            >
                                {node.condition && (OPS[node.condition.type] || []).map(o => <MenuItem key={o.value} value={o.value}>{o.label}</MenuItem>)}
                            </TextField>
                        </Grid>
                        <Grid size={{ xs: 12, md: 6 }}>
                            {node.condition?.type === 'User' && (node.condition?.op === 'in_group' || node.condition?.op === 'is') ? (
                                <Autocomplete
                                    freeSolo
                                    size="small"
                                    options={identityOptions}
                                    getOptionLabel={(option) => typeof option === 'string' ? option : option.label}
                                    value={node.condition.value}
                                    onInputChange={(_, val) => handleSearchIdentity(val)}
                                    onChange={(_, val) => {
                                        const finalValue = typeof val === 'string' ? val : (val?.value || '');
                                        onChange({ ...node, condition: { ...node.condition!, value: finalValue } });
                                    }}
                                    loading={searching}
                                    renderInput={(params) => (
                                        <TextField
                                            {...params}
                                            label={node.condition?.op === 'in_group' ? "Group" : "User Email"}
                                            placeholder="Search Google Identity..."
                                            sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                                            InputProps={{
                                                ...params.InputProps,
                                                endAdornment: (
                                                    <React.Fragment>
                                                        {searching ? <CircularProgress color="inherit" size={20} /> : null}
                                                        {params.InputProps.endAdornment}
                                                    </React.Fragment>
                                                ),
                                            }}
                                        />
                                    )}
                                    renderOption={(props, option) => (
                                        <MenuItem {...props} sx={{ gap: 1 }}>
                                            {option.type === 'user' ? <GroupIcon fontSize="small" color="action" /> : <AddIcon fontSize="small" color="action" />}
                                            <Box>
                                                <Typography variant="body2">{option.label}</Typography>
                                                <Typography variant="caption" color="text.secondary">{option.type.toUpperCase()}</Typography>
                                            </Box>
                                        </MenuItem>
                                    )}
                                />
                            ) : node.condition?.type === 'Network' && node.condition?.op === 'country' ? (
                                <Autocomplete
                                    size="small"
                                    options={COUNTRIES}
                                    getOptionLabel={(option) => option.label}
                                    value={COUNTRIES.find(c => c.code === node.condition?.value) || null}
                                    onChange={(_, val) => onChange({ ...node, condition: { ...node.condition!, value: val?.code || '' } })}
                                    renderInput={(params) => (
                                        <TextField
                                            {...params}
                                            label="Country"
                                            placeholder="Search country..."
                                            sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                                        />
                                    )}
                                />
                            ) : node.condition?.type === 'Network' && node.condition?.op === 'is_private' ? (
                                <Box sx={{ p: 1, bgcolor: 'rgba(0,0,0,0.03)', borderRadius: 2, display: 'flex', alignItems: 'center', height: '40px' }}>
                                    <Typography variant="body2" color="text.secondary">
                                        Matches any Private IP (RFC 1918)
                                    </Typography>
                                </Box>
                            ) : node.condition?.type === 'Device' && node.condition?.op === 'os' ? (
                                <Autocomplete
                                    size="small"
                                    options={OS_OPTIONS}
                                    getOptionLabel={(option) => option.label}
                                    value={OS_OPTIONS.find(o => o.value === node.condition?.value) || null}
                                    onChange={(_, val) => onChange({ ...node, condition: { ...node.condition!, value: val?.value || '' } })}
                                    renderInput={(params) => (
                                        <TextField
                                            {...params}
                                            label="Operating System"
                                            placeholder="Select OS..."
                                            sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                                        />
                                    )}
                                />
                            ) : (
                                <TextField
                                    fullWidth
                                    size="small"
                                    label="Value"
                                    value={node.condition?.value}
                                    onChange={(e) => onChange({ ...node, condition: { ...node.condition!, value: e.target.value } })}
                                    placeholder={node.condition?.type === 'User' ? 'e.g. user@domain.com' : 'e.g. 192.168.1.0/24'}
                                    sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                                />
                            )}
                        </Grid>
                        <Grid size={{ xs: 12, md: 1 }} sx={{ textAlign: 'right' }}>
                            {onDelete && (
                                <IconButton size="small" color="error" onClick={onDelete} sx={{ '&:hover': { bgcolor: 'rgba(211, 47, 47, 0.04)' } }}>
                                    <DeleteIcon fontSize="small" />
                                </IconButton>
                            )}
                        </Grid>
                    </Grid>
                )}

                {!isLeaf && <Box sx={{ flexGrow: 1 }} />}

                {!isLeaf && onDelete && (
                    <IconButton size="small" color="error" onClick={onDelete}>
                        <DeleteIcon fontSize="small" />
                    </IconButton>
                )}
            </Box>

            {
                !isLeaf && (
                    <Box sx={{ ml: depth > 0 ? 1 : 0, pl: depth > 0 ? 2 : 0, borderLeft: depth > 0 ? '1px dashed #ddd' : 'none' }}>
                        {(node.children || []).map((child, idx) => (
                            <NodeEditor
                                key={idx}
                                node={child}
                                stage={stage}
                                depth={depth + 1}
                                onChange={(updated) => handleChildChange(idx, updated)}
                                onDelete={() => handleRemoveChild(idx)}
                            />
                        ))}
                        <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                            <Button
                                size="small"
                                startIcon={<AddIcon />}
                                onClick={handleAddCondition}
                                variant="text"
                                sx={{
                                    fontSize: '0.75rem',
                                    fontWeight: 700,
                                    color: '#1a73e8',
                                    borderRadius: 2,
                                    '&:hover': { bgcolor: 'rgba(26, 115, 232, 0.04)' }
                                }}
                            >
                                Add Condition
                            </Button>
                            <Button
                                size="small"
                                startIcon={<AddIcon />}
                                onClick={handleAddSubGroup}
                                variant="text"
                                color="secondary"
                                sx={{
                                    fontSize: '0.75rem',
                                    fontWeight: 700,
                                    borderRadius: 2
                                }}
                            >
                                Add Sub-Group
                            </Button>
                        </Stack>
                    </Box>
                )
            }
        </Paper >
    );
};

const SignInPoliciesView: React.FC<SignInPoliciesViewProps> = ({ policies, onRefresh }) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const isSmallMobile = useMediaQuery(theme.breakpoints.down('sm'));

    const [activeTab, setActiveTab] = useState<'pre_auth' | 'post_auth'>('pre_auth');
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingPolicy, setEditingPolicy] = useState<SignInPolicy | null>(null);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState<{
        name: string;
        priority: number;
        block: boolean;
        stage: 'pre_auth' | 'post_auth';
        root_node: PolicyNode;
    }>({
        name: '',
        priority: 100,
        block: false,
        stage: 'pre_auth',
        root_node: { operator: 'AND', children: [] }
    });

    // Pagination State
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);

    const handleChangePage = (event: unknown, newPage: number) => {
        setPage(newPage);
    };

    const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
        setRowsPerPage(parseInt(event.target.value, 10));
        setPage(0);
    };

    const [confirmDialog, setConfirmDialog] = useState<{
        open: boolean;
        title: string;
        message: string;
        onConfirm: () => void;
    }>({
        open: false,
        title: '',
        message: '',
        onConfirm: () => { },
    });

    const handleOpenDialog = (policy?: SignInPolicy) => {
        if (policy) {
            setEditingPolicy(policy);
            setFormData({
                name: policy.name,
                priority: policy.priority,
                block: policy.block,
                stage: policy.stage || 'pre_auth',
                root_node: policy.root_node || { operator: 'AND', children: [] }
            });
        } else {
            setEditingPolicy(null);
            setFormData({
                name: '',
                priority: 100,
                block: false,
                stage: activeTab,
                root_node: { operator: 'AND', children: [] }
            });
        }
        setDialogOpen(true);
    };

    const handleCloseDialog = () => {
        setDialogOpen(false);
        setEditingPolicy(null);
    };

    const handleSubmit = async () => {
        setLoading(true);
        const url = '/api/v1/policies/sign-in';
        const method = editingPolicy ? 'PATCH' : 'POST';
        // Explicitly include enabled status to prevent resetting it to false (go zero value) on update
        const body = editingPolicy
            ? { ...formData, id: editingPolicy.id, enabled: editingPolicy.enabled }
            : { ...formData, enabled: true };

        try {
            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
            const data = await res.json();
            if (data.success) {
                onRefresh();
                handleCloseDialog();
            } else {
                alert('Error: ' + (data.error || data.message));
            }
        } catch (err) {
            console.error('Failed to save policy:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = (id: string) => {
        setConfirmDialog({
            open: true,
            title: 'Delete Policy',
            message: 'Are you sure you want to delete this policy?',
            onConfirm: async () => {
                try {
                    const res = await fetch(`/api/v1/policies/sign-in?id=${id}`, {
                        method: 'DELETE'
                    });
                    if (res.ok) {
                        onRefresh();
                    } else {
                        const data = await res.json();
                        alert('Error: ' + (data.error || data.message));
                    }
                } catch (err) {
                    console.error('Failed to delete policy:', err);
                }
                setConfirmDialog(prev => ({ ...prev, open: false }));
            }
        });
    };

    const handleToggleEnabled = (policy: SignInPolicy) => {
        const isEnabled = policy.enabled !== false;
        const action = isEnabled ? 'disable' : 'enable';

        setConfirmDialog({
            open: true,
            title: `${isEnabled ? 'Disable' : 'Enable'} Policy`,
            message: `Are you sure you want to ${action} this policy?`,
            onConfirm: async () => {
                try {
                    const res = await fetch('/api/v1/policies/sign-in', {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ ...policy, enabled: !isEnabled })
                    });
                    const data = await res.json();
                    if (data.success) {
                        onRefresh();
                    } else {
                        alert('Error: ' + (data.error || 'Failed to update policy'));
                    }
                } catch (err) {
                    console.error('Failed to toggle policy:', err);
                }
                setConfirmDialog(prev => ({ ...prev, open: false }));
            }
        });
    };

    const renderNodeSummary = (node: PolicyNode): React.ReactNode => {
        if (node.condition) {
            return (
                <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.5 }}>
                    <Chip
                        label={node.condition.type}
                        size="small"
                        sx={{ height: 18, fontSize: '0.65rem', fontWeight: 800, bgcolor: 'rgba(0,0,0,0.05)', color: '#5f6368', borderRadius: 1 }}
                    />
                    <Typography component="span" variant="caption" sx={{ fontWeight: 700, color: '#1a73e8', mx: 0.5 }}>
                        {node.condition.op.replace('_', ' ')}
                    </Typography>
                    <Chip
                        label={node.condition.value}
                        size="small"
                        variant="outlined"
                        sx={{ height: 18, fontSize: '0.65rem', fontWeight: 800, color: '#202124', borderRadius: 1, borderColor: '#dadce0' }}
                    />
                </Box>
            );
        }
        if (!node.children || node.children.length === 0) return <Typography component="span" variant="caption" sx={{ fontWeight: 700 }}>TRUE</Typography>;

        return (
            <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
                <Typography component="span" variant="caption" sx={{ fontWeight: 800, color: node.operator === 'AND' ? '#1a73e8' : '#f4b400', fontSize: '0.6rem' }}>
                    (
                </Typography>
                {node.children.map((c, idx) => (
                    <React.Fragment key={idx}>
                        {idx > 0 && (
                            <Typography component="span" variant="caption" sx={{ fontWeight: 900, color: node.operator === 'AND' ? '#1a73e8' : '#f4b400', px: 0.5, fontSize: '0.65rem' }}>
                                {node.operator}
                            </Typography>
                        )}
                        {renderNodeSummary(c)}
                    </React.Fragment>
                ))}
                <Typography component="span" variant="caption" sx={{ fontWeight: 800, color: node.operator === 'AND' ? '#1a73e8' : '#f4b400', fontSize: '0.6rem' }}>
                    )
                </Typography>
            </Box>
        );
    };

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant="h4" sx={{ fontWeight: 800, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 2 }}>
                        <SecurityIcon sx={{ fontSize: 40, color: 'primary.main' }} />
                        Sign-in Policies
                    </Typography>
                    <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
                        Implement conditional access and zero-trust security policies for user authentication.
                    </Typography>
                </Box>
                <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={() => handleOpenDialog()}
                    sx={{
                        borderRadius: 2.5,
                        px: 3,
                        py: 1,
                        boxShadow: '0 4px 12px rgba(26, 115, 232, 0.2)',
                        textTransform: 'none',
                        fontWeight: 700
                    }}
                >
                    Create Sign-in Policy
                </Button>
            </Box>

            <Box sx={{ mb: 4 }}>
                <ToggleButtonGroup
                    value={activeTab}
                    exclusive
                    onChange={(_, val) => val && setActiveTab(val)}
                    color="primary"
                    sx={{
                        width: '100%',
                        bgcolor: '#fff',
                        borderRadius: 3,
                        p: 0.5,
                        boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
                        '& .MuiToggleButton-root': {
                            flex: 1,
                            borderRadius: 2.5,
                            py: 1.5,
                            border: 'none',
                            textTransform: 'none',
                            fontWeight: 700,
                            '&.Mui-selected': {
                                bgcolor: 'primary.soft',
                                color: 'primary.main',
                                '&:hover': { bgcolor: 'primary.soft' }
                            }
                        }
                    }}
                >
                    <ToggleButton value="pre_auth">
                        Pre-authenticated Policies
                    </ToggleButton>
                    <ToggleButton value="post_auth">
                        Post-authenticated Policies
                    </ToggleButton>
                </ToggleButtonGroup>
            </Box>

            <Box sx={{ mb: 3 }}>
                <Typography variant="body2" color="text.secondary" sx={{ fontWeight: 500 }}>
                    {activeTab === 'pre_auth'
                        ? 'Evaluated BEFORE user authentication (e.g. Network/IP, Device checks).'
                        : 'Evaluated AFTER user authentication (e.g. User Identity, Group membership).'}
                </Typography>
            </Box>

            <Grid container spacing={3}>
                {policies.filter(p => (p.stage || 'pre_auth') === activeTab).length === 0 ? (
                    <Grid size={12}>
                        <Paper sx={{ p: 8, textAlign: 'center', borderRadius: 4, bgcolor: 'grey.50', border: '2px dashed #e0e0e0' }}>
                            <WarningIcon sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
                            <Typography variant="h6" color="text.secondary">No {activeTab.replace('_', '-')} policies defined</Typography>
                            <Typography variant="body2" color="text.disabled" sx={{ mt: 1 }}>
                                Start by creating your first policy for this stage.
                            </Typography>
                        </Paper>
                    </Grid>
                ) : (
                    policies.filter(p => (p.stage || 'pre_auth') === activeTab).sort((a, b) => a.priority - b.priority).slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((policy) => (
                        <Grid key={policy.id} size={12}>
                            <Card sx={{
                                borderRadius: 4,
                                border: '1px solid #eef0f2',
                                boxShadow: '0 2px 8px rgba(0,0,0,0.03)',
                                transition: 'all 0.2s',
                                opacity: policy.enabled === false ? 0.7 : 1,
                                bgcolor: '#fff',
                                '&:hover': {
                                    boxShadow: '0 8px 16px rgba(0,0,0,0.06)',
                                    borderColor: 'primary.main',
                                    transform: 'translateY(-2px)'
                                }
                            }}>
                                <Box sx={{ p: isSmallMobile ? 2 : 3 }}>
                                    <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', mb: 2 }}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                            <Tooltip title={policy.block ? "Deny Sign-in" : "Allow Sign-in"}>
                                                <Box sx={{
                                                    width: 40,
                                                    height: 40,
                                                    borderRadius: '50%',
                                                    bgcolor: policy.block ? 'error.soft' : 'success.soft',
                                                    color: policy.block ? 'error.main' : 'success.main',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                    flexShrink: 0,
                                                    border: `1px solid ${policy.block ? '#fad4d4' : '#b4f0b8'}`
                                                }}>
                                                    {policy.block ? <BlockIcon /> : <CheckCircleIcon />}
                                                </Box>
                                            </Tooltip>
                                            <Box>
                                                <Typography variant="h6" sx={{ fontWeight: 800, fontSize: isSmallMobile ? '1rem' : '1.1rem', lineHeight: 1.2 }}>
                                                    {policy.name}
                                                </Typography>
                                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                                                    <Chip
                                                        label={policy.enabled !== false ? "Active" : "Disabled"}
                                                        size="small"
                                                        color={policy.enabled !== false ? "success" : "default"}
                                                        sx={{ height: 20, fontSize: '0.65rem', fontWeight: 700, borderRadius: 1 }}
                                                    />
                                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600 }}>
                                                        Priority: {policy.priority}
                                                    </Typography>
                                                </Box>
                                            </Box>
                                        </Box>

                                        <Box sx={{ display: 'flex', gap: 1 }}>
                                            {!isSmallMobile && (
                                                <>
                                                    <Tooltip title={policy.enabled !== false ? "Disable" : "Enable"}>
                                                        <IconButton
                                                            size="small"
                                                            onClick={() => handleToggleEnabled(policy)}
                                                            sx={{ color: 'text.secondary', border: '1px solid #e0e0e0', borderRadius: 2 }}
                                                        >
                                                            {policy.enabled !== false ? <PauseIcon fontSize="small" /> : <PlayArrowIcon fontSize="small" />}
                                                        </IconButton>
                                                    </Tooltip>
                                                    <Tooltip title="Edit Policy">
                                                        <IconButton
                                                            size="small"
                                                            onClick={() => handleOpenDialog(policy)}
                                                            sx={{ color: 'primary.main', border: '1px solid #e0e0e0', bgcolor: 'primary.50', borderRadius: 2 }}
                                                        >
                                                            <EditIcon fontSize="small" />
                                                        </IconButton>
                                                    </Tooltip>
                                                </>
                                            )}
                                            <Tooltip title="Delete Policy">
                                                <IconButton
                                                    size="small"
                                                    onClick={() => handleDelete(policy.id)}
                                                    sx={{ color: 'error.main', border: '1px solid #e0e0e0', bgcolor: 'error.50', borderRadius: 2 }}
                                                >
                                                    <DeleteIcon fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                        </Box>
                                    </Box>

                                    <Divider sx={{ my: 2, borderStyle: 'dashed' }} />

                                    <Grid container spacing={3}>
                                        <Grid item xs={12} md={9}>
                                            <Box>
                                                <Typography variant="caption" sx={{ textTransform: 'uppercase', color: 'text.secondary', fontWeight: 700, display: 'flex', alignItems: 'center', gap: 0.5, mb: 1 }}>
                                                    <GroupIcon fontSize="inherit" /> Conditions
                                                </Typography>
                                                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                                                    {policy.root_node ? renderNodeSummary(policy.root_node) : <Typography variant="body2" color="text.secondary">No conditions set (True)</Typography>}
                                                </Box>
                                            </Box>
                                        </Grid>
                                        <Grid item xs={12} md={3}>
                                            <Box>
                                                <Typography variant="caption" sx={{ textTransform: 'uppercase', color: 'text.secondary', fontWeight: 700, display: 'flex', alignItems: 'center', gap: 0.5, mb: 1 }}>
                                                    <SecurityIcon fontSize="inherit" /> Action
                                                </Typography>
                                                <Chip
                                                    label={policy.block ? 'BLOCK SIGN-IN' : 'ALLOW SIGN-IN'}
                                                    size="small"
                                                    sx={{
                                                        fontWeight: 800,
                                                        bgcolor: policy.block ? '#fce8e6' : '#e6f4ea',
                                                        color: policy.block ? '#c5221f' : '#137333',
                                                        borderRadius: 1
                                                    }}
                                                />
                                            </Box>
                                        </Grid>
                                    </Grid>

                                    {isSmallMobile && (
                                        <Box sx={{ display: 'flex', gap: 1, mt: 3, pt: 2, borderTop: '1px solid #f1f3f4' }}>
                                            <Button
                                                size="small"
                                                fullWidth
                                                variant="outlined"
                                                onClick={() => handleToggleEnabled(policy)}
                                                startIcon={policy.enabled !== false ? <PauseIcon /> : <PlayArrowIcon />}
                                            >
                                                {policy.enabled !== false ? "Disable" : "Enable"}
                                            </Button>
                                            <Button
                                                size="small"
                                                fullWidth
                                                variant="outlined"
                                                onClick={() => handleOpenDialog(policy)}
                                                startIcon={<EditIcon />}
                                            >
                                                Edit
                                            </Button>
                                        </Box>
                                    )}
                                </Box>
                            </Card>
                        </Grid>
                    ))
                )}
            </Grid>

            {policies.filter(p => (p.stage || 'pre_auth') === activeTab).length > 0 && (
                <TablePagination
                    component="div"
                    count={policies.filter(p => (p.stage || 'pre_auth') === activeTab).length}
                    page={page}
                    onPageChange={handleChangePage}
                    rowsPerPage={rowsPerPage}
                    onRowsPerPageChange={handleChangeRowsPerPage}
                    labelRowsPerPage="Policies per page:"
                    sx={{ mt: 2 }}
                />
            )}



            <Dialog
                open={dialogOpen}
                fullScreen={isMobile}
                onClose={(_, reason) => {
                    if (reason !== 'backdropClick') handleCloseDialog();
                }}
                maxWidth="md"
                fullWidth
                PaperProps={{
                    sx: { borderRadius: isMobile ? 0 : 4, minHeight: isMobile ? '100vh' : 'auto', boxShadow: '0 24px 48px rgba(0,0,0,0.1)' }
                }}
            >
                <DialogTitle sx={{ px: 4, pt: 4, pb: 2 }}>
                    <Typography variant="h5" sx={{ fontWeight: 800 }}>
                        {editingPolicy ? 'Edit Sign-in Policy' : 'Create Sign-in Policy'}
                    </Typography>
                </DialogTitle>
                <DialogContent sx={{ px: 4 }}>
                    <Grid container spacing={3} sx={{ mb: 4, pt: 3 }}>
                        <Grid size={{ xs: 12, md: 6 }}>
                            <TextField
                                fullWidth
                                variant="outlined"
                                label="Policy Name"
                                placeholder="e.g. Block login from outside TH"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            />
                        </Grid>
                        <Grid size={{ xs: 12, md: 3 }}>
                            <TextField
                                select
                                fullWidth
                                label="Policy Stage"
                                value={formData.stage}
                                onChange={(e) => setFormData({ ...formData, stage: e.target.value as any })}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            >
                                <MenuItem value="pre_auth">Pre-auth</MenuItem>
                                <MenuItem value="post_auth">Post-auth</MenuItem>
                            </TextField>
                        </Grid>
                        <Grid size={{ xs: 12, md: 3 }}>
                            <TextField
                                fullWidth
                                type="number"
                                label="Priority"
                                value={formData.priority}
                                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                                helperText="Lower is higher precedence"
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            />
                        </Grid>
                    </Grid>

                    <Typography variant="subtitle2" sx={{ fontWeight: 700, color: '#5f6368', mb: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Box sx={{ width: 4, height: 16, bgcolor: '#1a73e8', borderRadius: 4 }} />
                        CONDITIONAL LOGIC
                    </Typography>

                    <NodeEditor
                        node={formData.root_node}
                        stage={formData.stage}
                        onChange={(node) => setFormData({ ...formData, root_node: node })}
                    />

                    <Grid size={12}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1, color: 'text.secondary', display: 'flex', alignItems: 'center', gap: 1 }}>
                            <SecurityIcon sx={{ fontSize: 16 }} /> POLICY ACTION (EFFECT)
                        </Typography>
                        <Box sx={{ display: 'flex', gap: 2 }}>
                            <Card
                                onClick={() => setFormData({ ...formData, block: false })}
                                sx={{
                                    flex: 1,
                                    cursor: 'pointer',
                                    borderRadius: 3,
                                    border: !formData.block ? '2px solid #34a853' : '1px solid #e0e0e0',
                                    bgcolor: !formData.block ? 'rgba(52, 168, 83, 0.04)' : '#fff',
                                    transition: 'all 0.2s',
                                    '&:hover': { borderColor: '#34a853', bgcolor: 'rgba(52, 168, 83, 0.02)' }
                                }}
                            >
                                <CardContent sx={{ textAlign: 'center', p: 2 }}>
                                    <CheckCircleIcon sx={{ color: !formData.block ? '#34a853' : '#ccc', fontSize: 32, mb: 1 }} />
                                    <Typography variant="subtitle1" sx={{ fontWeight: 800, color: !formData.block ? '#34a853' : '#666' }}>ALLOW SIGN-IN</Typography>
                                </CardContent>
                            </Card>
                            <Card
                                onClick={() => setFormData({ ...formData, block: true })}
                                sx={{
                                    flex: 1,
                                    cursor: 'pointer',
                                    borderRadius: 3,
                                    border: formData.block ? '2px solid #ea4335' : '1px solid #e0e0e0',
                                    bgcolor: formData.block ? 'rgba(234, 67, 53, 0.04)' : '#fff',
                                    transition: 'all 0.2s',
                                    '&:hover': { borderColor: '#ea4335', bgcolor: 'rgba(234, 67, 53, 0.02)' }
                                }}
                            >
                                <CardContent sx={{ textAlign: 'center', p: 2 }}>
                                    <BlockIcon sx={{ color: formData.block ? '#ea4335' : '#ccc', fontSize: 32, mb: 1 }} />
                                    <Typography variant="subtitle1" sx={{ fontWeight: 800, color: formData.block ? '#ea4335' : '#666' }}>BLOCK SIGN-IN</Typography>
                                </CardContent>
                            </Card>
                        </Box>
                    </Grid>
                </DialogContent>
                <DialogActions sx={{ px: 4, pb: 4, pt: 2 }}>
                    <Button onClick={handleCloseDialog} sx={{ fontWeight: 700 }}>Cancel</Button>
                    <Button
                        variant="contained"
                        onClick={handleSubmit}
                        disabled={loading}
                        sx={{
                            borderRadius: 2.5,
                            px: 4,
                            fontWeight: 700,
                            boxShadow: '0 4px 12px rgba(26, 115, 232, 0.2)'
                        }}
                    >
                        {loading ? <CircularProgress size={24} color="inherit" /> : (editingPolicy ? 'Update Policy' : 'Create Policy')}
                    </Button>
                </DialogActions>
            </Dialog>

            <Dialog
                open={confirmDialog.open}
                onClose={() => setConfirmDialog({ ...confirmDialog, open: false })}
                fullWidth
                maxWidth="xs"
                PaperProps={{ sx: { borderRadius: 3, p: 1 } }}
            >
                <DialogTitle sx={{ fontWeight: 800 }}>{confirmDialog.title}</DialogTitle>
                <DialogContent>
                    <Typography>{confirmDialog.message}</Typography>
                </DialogContent>
                <DialogActions sx={{ p: 2 }}>
                    <Button onClick={() => setConfirmDialog({ ...confirmDialog, open: false })} color="inherit" sx={{ fontWeight: 700 }}>Cancel</Button>
                    <Button onClick={confirmDialog.onConfirm} variant="contained" color="primary" sx={{ fontWeight: 700, px: 3, borderRadius: 2 }}>Confirm</Button>
                </DialogActions>
            </Dialog>
        </Box >
    );
};

export default SignInPoliciesView;
