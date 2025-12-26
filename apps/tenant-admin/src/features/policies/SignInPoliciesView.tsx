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
    createFilterOptions
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
    KeyboardArrowRight as ArrowIcon
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

const OPS: Record<string, { label: string; value: string }[]> = {
    'User': [{ label: 'In Group', value: 'in_group' }, { label: 'Email Ends With', value: 'email_suffix' }],
    'Network': [{ label: 'IP in CIDR', value: 'cidr' }, { label: 'Country Equals', value: 'country' }],
    'Device': [{ label: 'OS Equals', value: 'os' }, { label: 'Is Managed', value: 'managed' }],
};

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
    onChange: (newNode: PolicyNode) => void;
    onDelete?: () => void;
    depth?: number;
}> = ({ node, onChange, onDelete, depth = 0 }) => {
    const isLeaf = !!node.condition;

    const handleAddCondition = () => {
        const newNode: PolicyNode = {
            operator: 'AND',
            condition: { type: 'User', field: 'group', op: 'in_group', value: '' }
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
                setIdentityOptions(data.data);
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
                        <Grid size={2.5}>
                            <TextField
                                select
                                fullWidth
                                size="small"
                                label="Type"
                                value={node.condition?.type}
                                onChange={(e) => onChange({ ...node, condition: { ...node.condition!, type: e.target.value, op: OPS[e.target.value][0].value, value: '' } })}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            >
                                {CONDITION_TYPES.map(t => <MenuItem key={t.value} value={t.value} sx={{ gap: 1 }}>{t.icon} {t.label}</MenuItem>)}
                            </TextField>
                        </Grid>
                        <Grid size={2.5}>
                            <TextField
                                select
                                fullWidth
                                size="small"
                                label="Operator"
                                value={node.condition?.op}
                                onChange={(e) => onChange({ ...node, condition: { ...node.condition!, op: e.target.value, value: '' } })}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 2 } }}
                            >
                                {node.condition && (OPS[node.condition.type] || []).map(o => <MenuItem key={o.value} value={o.value}>{o.label}</MenuItem>)}
                            </TextField>
                        </Grid>
                        <Grid size={6}>
                            {node.condition?.type === 'User' && node.condition?.op === 'in_group' ? (
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
                                            label="User / Group"
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
                        <Grid size={1} sx={{ textAlign: 'right' }}>
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

            {!isLeaf && (
                <Box sx={{ ml: depth > 0 ? 1 : 0, pl: depth > 0 ? 2 : 0, borderLeft: depth > 0 ? '1px dashed #ddd' : 'none' }}>
                    {(node.children || []).map((child, idx) => (
                        <NodeEditor
                            key={idx}
                            node={child}
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
            )}
        </Paper>
    );
};

const SignInPoliciesView: React.FC<SignInPoliciesViewProps> = ({ policies, onRefresh }) => {
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingPolicy, setEditingPolicy] = useState<SignInPolicy | null>(null);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState<{
        name: string;
        priority: number;
        block: boolean;
        root_node: PolicyNode;
    }>({
        name: '',
        priority: 10,
        block: false,
        root_node: { operator: 'AND', children: [] }
    });

    const handleOpenDialog = (policy?: SignInPolicy) => {
        if (policy) {
            setEditingPolicy(policy);
            setFormData({
                name: policy.name,
                priority: policy.priority,
                block: policy.block,
                root_node: policy.root_node || { operator: 'AND', children: [] }
            });
        } else {
            setEditingPolicy(null);
            setFormData({
                name: '',
                priority: 10,
                block: false,
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
        const body = editingPolicy ? { ...formData, id: editingPolicy.id } : formData;

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

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this policy?')) return;
        try {
            const res = await fetch('/api/v1/policies/sign-in', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id })
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
        <Box sx={{ maxWidth: 1200, mx: 'auto', py: 4 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant="h4" sx={{ fontWeight: 800, color: '#202124' }}>Sign-in Policies</Typography>
                    <Typography color="text.secondary">Implement conditional access and zero-trust security policies.</Typography>
                </Box>
                <Button
                    variant="contained"
                    disableElevation
                    startIcon={<AddIcon />}
                    onClick={() => handleOpenDialog()}
                    sx={{ borderRadius: 2, px: 3, bgcolor: '#1a73e8', '&:hover': { bgcolor: '#1765cc' } }}
                >
                    Add Policy
                </Button>
            </Box>

            {policies.length === 0 ? (
                <Paper variant="outlined" sx={{ p: 8, textAlign: 'center', borderRadius: 4, bgcolor: '#fff', border: '1px dashed #dadce0' }}>
                    <WarningIcon sx={{ fontSize: 48, color: '#dadce0', mb: 2 }} />
                    <Typography variant="h6" sx={{ fontWeight: 600, color: '#3c4043' }}>No policies configured</Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 4, maxWidth: 400, mx: 'auto' }}>
                        Zero Trust policies allow you to control access based on user context, location, and device status.
                    </Typography>
                    <Button variant="outlined" startIcon={<AddIcon />} onClick={() => handleOpenDialog()} sx={{ borderRadius: 2 }}>
                        Create First Policy
                    </Button>
                </Paper>
            ) : (
                <Stack spacing={2}>
                    {policies.sort((a, b) => a.priority - b.priority).map((policy) => (
                        <Card key={policy.id} variant="outlined" sx={{ borderRadius: 3, border: '1px solid #dadce0', transition: '0.2s', '&:hover': { boxShadow: '0 1px 6px rgba(32,33,36,.28)', borderColor: 'transparent' } }}>
                            <CardContent sx={{ p: 2.5 }}>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2.5 }}>
                                    <Box sx={{
                                        width: 48,
                                        height: 48,
                                        borderRadius: '50%',
                                        bgcolor: policy.block ? 'rgba(217, 48, 37, 0.1)' : 'rgba(30, 142, 62, 0.1)',
                                        color: policy.block ? '#d93025' : '#1e8e32',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center'
                                    }}>
                                        {policy.block ? <BlockIcon /> : <CheckCircleIcon />}
                                    </Box>
                                    <Box sx={{ flexGrow: 1 }}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 0.5 }}>
                                            <Typography variant="subtitle1" sx={{ fontWeight: 700, color: '#202124' }}>{policy.name}</Typography>
                                            <Chip
                                                label={`P${policy.priority}`}
                                                size="small"
                                                variant="outlined"
                                                sx={{ height: 20, fontSize: 10, fontWeight: 700, color: '#5f6368', borderColor: '#dadce0' }}
                                            />
                                            <Chip
                                                label={policy.block ? 'DENY' : 'ALLOW'}
                                                size="small"
                                                sx={{
                                                    height: 20,
                                                    fontSize: 10,
                                                    fontWeight: 900,
                                                    bgcolor: policy.block ? '#fce8e6' : '#e6f4ea',
                                                    color: policy.block ? '#c5221f' : '#137333',
                                                    border: 'none'
                                                }}
                                            />
                                        </Box>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                            <Typography variant="body2" sx={{ 
                                                fontWeight: 800, 
                                                color: '#1a73e8', 
                                                fontSize: '0.7rem',
                                                bgcolor: 'rgba(26, 115, 232, 0.08)',
                                                px: 1,
                                                py: 0.25,
                                                borderRadius: 1
                                            }}>IF</Typography>
                                            <Box sx={{ 
                                                display: 'flex', 
                                                alignItems: 'center', 
                                                gap: 1,
                                                bgcolor: '#f8f9fa',
                                                px: 1.5,
                                                py: 1,
                                                borderRadius: 2,
                                                border: '1px solid #f1f3f4',
                                                flexGrow: 1
                                            }}>
                                                {policy.root_node ? renderNodeSummary(policy.root_node) : <Typography variant="caption" sx={{ fontWeight: 700 }}>TRUE</Typography>}
                                            </Box>
                                        </Box>
                                    </Box>
                                    <Box sx={{ display: 'flex', gap: 1 }}>
                                        <Tooltip title="Edit Policy">
                                            <IconButton onClick={() => handleOpenDialog(policy)} size="small" sx={{ color: '#5f6368' }}>
                                                <EditIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="Delete Policy">
                                            <IconButton onClick={() => handleDelete(policy.id)} size="small" color="error">
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                    </Box>
                                </Box>
                            </CardContent>
                        </Card>
                    ))}
                </Stack>
            )}

            <Dialog
                open={dialogOpen}
                onClose={handleCloseDialog}
                maxWidth="md"
                fullWidth
                PaperProps={{
                    sx: { borderRadius: 4, boxShadow: '0 24px 38px 3px rgba(0,0,0,0.14), 0 9px 46px 8px rgba(0,0,0,0.12), 0 11px 15px -7px rgba(0,0,0,0.2)' }
                }}
            >
                <DialogTitle sx={{ fontWeight: 800, p: 3, color: '#202124', borderBottom: '1px solid #f1f3f4' }}>
                    {editingPolicy ? 'Edit Access Policy' : 'Create Zero Trust Policy'}
                </DialogTitle>
                <DialogContent sx={{ p: 4, pt: 3 }}>
                    <Grid container spacing={3} sx={{ mb: 4, pt:3 }}>
                        <Grid size={8}>
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
                        <Grid size={4}>
                            <TextField
                                fullWidth
                                type="number"
                                label="Priority"
                                value={formData.priority}
                                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                                helperText="Lower values have higher precedence"
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
                        onChange={(node) => setFormData({ ...formData, root_node: node })}
                    />

                    <Box sx={{ mt: 4 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 2, color: '#5f6368', display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Box sx={{ width: 4, height: 16, bgcolor: formData.block ? '#d93025' : '#1e8e3e', borderRadius: 4 }} />
                            POLICY ACTION
                        </Typography>
                        <Grid container spacing={2}>
                            <Grid size={6}>
                                <Paper
                                    variant="outlined"
                                    onClick={() => setFormData({ ...formData, block: false })}
                                    sx={{
                                        p: 2.5,
                                        cursor: 'pointer',
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 2,
                                        borderRadius: 3,
                                        borderWidth: 2,
                                        borderColor: !formData.block ? '#1e8e3e' : '#f1f3f4',
                                        bgcolor: !formData.block ? 'rgba(30, 142, 62, 0.04)' : 'transparent',
                                        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                                        '&:hover': { 
                                            bgcolor: !formData.block ? 'rgba(30, 142, 62, 0.08)' : '#f8f9fa',
                                            transform: 'translateY(-2px)',
                                            boxShadow: '0 4px 12px rgba(0,0,0,0.05)'
                                        }
                                    }}
                                >
                                    <Box sx={{
                                        width: 44,
                                        height: 44,
                                        borderRadius: '12px',
                                        bgcolor: !formData.block ? '#1e8e3e' : '#f1f3f4',
                                        color: !formData.block ? '#fff' : '#5f6368',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        transition: 'all 0.2s'
                                    }}>
                                        <CheckCircleIcon />
                                    </Box>
                                    <Box>
                                        <Typography sx={{ fontWeight: 800, fontSize: '0.9rem', color: !formData.block ? '#137333' : '#202124' }}>ALLOW ACCESS</Typography>
                                        <Typography variant="caption" sx={{ color: '#5f6368', display: 'block', lineHeight: 1.2 }}>Grant access if criteria met</Typography>
                                    </Box>
                                </Paper>
                            </Grid>
                            <Grid size={6}>
                                <Paper
                                    variant="outlined"
                                    onClick={() => setFormData({ ...formData, block: true })}
                                    sx={{
                                        p: 2.5,
                                        cursor: 'pointer',
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 2,
                                        borderRadius: 3,
                                        borderWidth: 2,
                                        borderColor: formData.block ? '#d93025' : '#f1f3f4',
                                        bgcolor: formData.block ? 'rgba(217, 48, 37, 0.04)' : 'transparent',
                                        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
                                        '&:hover': { 
                                            bgcolor: formData.block ? 'rgba(217, 48, 37, 0.08)' : '#f8f9fa',
                                            transform: 'translateY(-2px)',
                                            boxShadow: '0 4px 12px rgba(0,0,0,0.05)'
                                        }
                                    }}
                                >
                                    <Box sx={{
                                        width: 44,
                                        height: 44,
                                        borderRadius: '12px',
                                        bgcolor: formData.block ? '#d93025' : '#f1f3f4',
                                        color: formData.block ? '#fff' : '#5f6368',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        transition: 'all 0.2s'
                                    }}>
                                        <BlockIcon />
                                    </Box>
                                    <Box>
                                        <Typography sx={{ fontWeight: 800, fontSize: '0.9rem', color: formData.block ? '#c5221f' : '#202124' }}>DENY ACCESS</Typography>
                                        <Typography variant="caption" sx={{ color: '#5f6368', display: 'block', lineHeight: 1.2 }}>Block access if criteria met</Typography>
                                    </Box>
                                </Paper>
                            </Grid>
                        </Grid>
                    </Box>
                </DialogContent>
                <DialogActions sx={{ p: 3, px: 4, bgcolor: '#f8f9fa', borderTop: '1px solid #f1f3f4' }}>
                    <Button onClick={handleCloseDialog} sx={{ color: '#5f6368', fontWeight: 600, px: 3 }}>Cancel</Button>
                    <Button
                        variant="contained"
                        disableElevation
                        onClick={handleSubmit}
                        disabled={!formData.name || loading}
                        sx={{ 
                            borderRadius: 2, 
                            px: 4, 
                            py: 1.2,
                            fontWeight: 700, 
                            bgcolor: '#1a73e8',
                            '&:hover': { bgcolor: '#1765cc' }
                        }}
                    >
                        {loading ? <CircularProgress size={20} color="inherit" /> : (editingPolicy ? 'Save Changes' : 'Create Policy')}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default SignInPoliciesView;
