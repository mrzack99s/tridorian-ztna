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
    DeviceHub as DeviceHubIcon,
    Public as PublicIcon,
    Group as GroupIcon,
    KeyboardArrowRight as ArrowIcon,
    Security as SecurityIcon,
    Dns as DnsIcon
} from '@mui/icons-material';
import { AccessPolicy, PolicyNode, PolicyCondition } from '../../types';

interface PoliciesViewProps {
    policies: AccessPolicy[];
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

const PoliciesView: React.FC<PoliciesViewProps> = ({ policies, onRefresh }) => {
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingPolicy, setEditingPolicy] = useState<AccessPolicy | null>(null);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState<{
        name: string;
        priority: number;
        effect: 'allow' | 'deny';
        destination: string;
        root_node: PolicyNode;
    }>({
        name: '',
        priority: 10,
        effect: 'allow',
        destination: '',
        root_node: { operator: 'AND', children: [] }
    });

    const handleOpenDialog = (policy?: AccessPolicy) => {
        if (policy) {
            setEditingPolicy(policy);
            setFormData({
                name: policy.name,
                priority: policy.priority,
                effect: policy.effect as any,
                destination: policy.destination || '',
                root_node: policy.root_node || { operator: 'AND', children: [] }
            });
        } else {
            setEditingPolicy(null);
            setFormData({
                name: '',
                priority: (policies.length + 1) * 10,
                effect: 'allow',
                destination: '',
                root_node: { operator: 'AND', children: [] }
            });
        }
        setDialogOpen(true);
    };

    const handleSave = async () => {
        setLoading(true);
        try {
            const url = '/api/v1/policies/access';
            const method = editingPolicy ? 'PATCH' : 'POST';
            const body = editingPolicy ? { ...formData, id: editingPolicy.id } : formData;

            const res = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                setDialogOpen(false);
                onRefresh();
            }
        } catch (err) {
            console.error('Save failed', err);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!window.confirm('Are you sure you want to delete this policy?')) return;
        try {
            const res = await fetch(`/api/v1/policies/access?id=${id}`, { method: 'DELETE' });
            if (res.ok) onRefresh();
        } catch (err) {
            console.error('Delete failed', err);
        }
    };

    const renderConditionSummary = (node: PolicyNode) => {
        if (node.condition) {
            return (
                <Chip
                    size="small"
                    label={`${node.condition.type}: ${node.condition.op} ${node.condition.value}`}
                    variant="outlined"
                    sx={{ mr: 0.5, mb: 0.5, fontSize: '0.7rem' }}
                />
            );
        }
        return (
            <Box sx={{ display: 'inline-flex', flexWrap: 'wrap', alignItems: 'center', gap: 0.5 }}>
                {node.children?.map((child, i) => (
                    <React.Fragment key={i}>
                        {i > 0 && <Typography variant="caption" sx={{ fontWeight: 'bold', color: 'text.secondary' }}>{node.operator}</Typography>}
                        {renderConditionSummary(child)}
                    </React.Fragment>
                ))}
            </Box>
        );
    };

    return (
        <Box sx={{ p: 0 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
                <Box>
                    <Typography variant="h4" sx={{ fontWeight: 800, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 2 }}>
                        <SecurityIcon sx={{ fontSize: 40, color: 'primary.main' }} />
                        Access Policies
                    </Typography>
                    <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
                        Define fine-grained access control rules for your network resources and services.
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
                    Create Access Policy
                </Button>
            </Box>

            <Grid container spacing={3}>
                {policies.map((policy) => (
                    <Grid key={policy.id} size={12}>
                        <Card sx={{
                            borderRadius: 4,
                            border: '1px solid #eef0f2',
                            boxShadow: '0 2px 8px rgba(0,0,0,0.03)',
                            transition: 'all 0.2s',
                            '&:hover': {
                                boxShadow: '0 8px 16px rgba(0,0,0,0.06)',
                                borderColor: 'primary.main'
                            }
                        }}>
                            <CardContent sx={{ p: 3 }}>
                                <Grid container spacing={2} alignItems="center">
                                    <Grid size={4}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                            <Box sx={{
                                                width: 48,
                                                height: 48,
                                                borderRadius: 3,
                                                bgcolor: policy.effect === 'allow' ? 'success.light' : 'error.light',
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                color: '#fff',
                                                opacity: 0.9
                                            }}>
                                                {policy.effect === 'allow' ? <CheckCircleIcon /> : <BlockIcon />}
                                            </Box>
                                            <Box>
                                                <Typography variant="h6" sx={{ fontWeight: 700 }}>{policy.name}</Typography>
                                                <Typography variant="caption" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                                    Priority: {policy.priority} • <DnsIcon sx={{ fontSize: 12 }} /> {policy.destination || 'All Destinations'}
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </Grid>
                                    <Grid size={6}>
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                            <Chip label="IF" size="small" sx={{ fontWeight: 900, bgcolor: 'grey.100', color: 'grey.700', borderRadius: 1 }} />
                                            <Box sx={{ display: 'flex', flexWrap: 'wrap' }}>
                                                {policy.root_node ? renderConditionSummary(policy.root_node) : <Typography variant="caption" color="text.secondary">No conditions (Always applies)</Typography>}
                                            </Box>
                                        </Box>
                                    </Grid>
                                    <Grid size={2} sx={{ textAlign: 'right' }}>
                                        <IconButton size="small" onClick={() => handleOpenDialog(policy)} sx={{ mr: 1, color: 'primary.main', bgcolor: 'rgba(26, 115, 232, 0.05)' }}>
                                            <EditIcon fontSize="small" />
                                        </IconButton>
                                        <IconButton size="small" onClick={() => handleDelete(policy.id)} sx={{ color: 'error.main', bgcolor: 'rgba(211, 47, 47, 0.05)' }}>
                                            <DeleteIcon fontSize="small" />
                                        </IconButton>
                                    </Grid>
                                </Grid>
                            </CardContent>
                        </Card>
                    </Grid>
                ))}

                {policies.length === 0 && (
                    <Grid size={12}>
                        <Paper sx={{ p: 8, textAlign: 'center', borderRadius: 4, bgcolor: 'grey.50', border: '2px dashed #e0e0e0' }}>
                            <SecurityIcon sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
                            <Typography variant="h6" color="text.secondary">No access policies defined</Typography>
                            <Typography variant="body2" color="text.disabled" sx={{ mt: 1 }}>
                                Start by creating your first policy to control access to network resources.
                            </Typography>
                        </Paper>
                    </Grid>
                )}
            </Grid>

            <Dialog
                open={dialogOpen}
                onClose={() => !loading && setDialogOpen(false)}
                maxWidth="md"
                fullWidth
                PaperProps={{
                    sx: { borderRadius: 4, boxShadow: '0 24px 48px rgba(0,0,0,0.1)' }
                }}
            >
                <DialogTitle sx={{ px: 4, pt: 4, pb: 2 }}>
                    <Typography variant="h5" sx={{ fontWeight: 800 }}>
                        {editingPolicy ? 'Edit Access Policy' : 'Create Access Policy'}
                    </Typography>
                </DialogTitle>
                <DialogContent sx={{ px: 4 }}>
                    <Grid container spacing={3} sx={{ mt: 0.5 }}>
                        <Grid size={8}>
                            <TextField
                                fullWidth
                                label="Policy Name"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="e.g. Sales Team Internal Access"
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                            />
                        </Grid>
                        <Grid size={4}>
                            <TextField
                                fullWidth
                                type="number"
                                label="Priority"
                                value={formData.priority}
                                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                            />
                        </Grid>

                        <Grid size={12}>
                            <TextField
                                fullWidth
                                label="Destination"
                                title="Target CIDR, IP, or Service name"
                                value={formData.destination}
                                onChange={(e) => setFormData({ ...formData, destination: e.target.value })}
                                placeholder="e.g. 10.0.0.0/24, internal-service.local, or admin-app"
                                sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                            />
                            <Typography variant="caption" color="text.secondary" sx={{ ml: 1, mt: 0.5, display: 'block' }}>
                                Specify the target network resource (CIDR, IP, or Service name) this policy applies to.
                            </Typography>
                        </Grid>

                        <Grid size={12}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1, color: 'text.secondary', display: 'flex', alignItems: 'center', gap: 1 }}>
                                <ArrowIcon sx={{ fontSize: 16 }} /> POLICY ACTION (EFFECT)
                            </Typography>
                            <Box sx={{ display: 'flex', gap: 2 }}>
                                <Card
                                    onClick={() => setFormData({ ...formData, effect: 'allow' })}
                                    sx={{
                                        flex: 1,
                                        cursor: 'pointer',
                                        borderRadius: 3,
                                        border: formData.effect === 'allow' ? '2px solid #34a853' : '1px solid #e0e0e0',
                                        bgcolor: formData.effect === 'allow' ? 'rgba(52, 168, 83, 0.04)' : '#fff',
                                        transition: 'all 0.2s',
                                        '&:hover': { borderColor: '#34a853', bgcolor: 'rgba(52, 168, 83, 0.02)' }
                                    }}
                                >
                                    <CardContent sx={{ textAlign: 'center', p: 2 }}>
                                        <CheckCircleIcon sx={{ color: formData.effect === 'allow' ? '#34a853' : '#ccc', fontSize: 32, mb: 1 }} />
                                        <Typography variant="subtitle1" sx={{ fontWeight: 800, color: formData.effect === 'allow' ? '#34a853' : '#666' }}>ALLOW ACCESS</Typography>
                                    </CardContent>
                                </Card>
                                <Card
                                    onClick={() => setFormData({ ...formData, effect: 'deny' })}
                                    sx={{
                                        flex: 1,
                                        cursor: 'pointer',
                                        borderRadius: 3,
                                        border: formData.effect === 'deny' ? '2px solid #ea4335' : '1px solid #e0e0e0',
                                        bgcolor: formData.effect === 'deny' ? 'rgba(234, 67, 53, 0.04)' : '#fff',
                                        transition: 'all 0.2s',
                                        '&:hover': { borderColor: '#ea4335', bgcolor: 'rgba(234, 67, 53, 0.02)' }
                                    }}
                                >
                                    <CardContent sx={{ textAlign: 'center', p: 2 }}>
                                        <BlockIcon sx={{ color: formData.effect === 'deny' ? '#ea4335' : '#ccc', fontSize: 32, mb: 1 }} />
                                        <Typography variant="subtitle1" sx={{ fontWeight: 800, color: formData.effect === 'deny' ? '#ea4335' : '#666' }}>DENY ACCESS</Typography>
                                    </CardContent>
                                </Card>
                            </Box>
                        </Grid>

                        <Grid size={12}>
                            <Divider sx={{ my: 1 }} />
                            <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 2, color: 'text.secondary', display: 'flex', alignItems: 'center', gap: 1 }}>
                                <ArrowIcon sx={{ fontSize: 16 }} /> CRITERIA (CONDITIONS)
                            </Typography>
                            <NodeEditor
                                node={formData.root_node}
                                onChange={(newNode) => setFormData({ ...formData, root_node: newNode })}
                            />
                        </Grid>
                    </Grid>
                </DialogContent>
                <DialogActions sx={{ px: 4, pb: 4, pt: 2 }}>
                    <Button onClick={() => setDialogOpen(false)} disabled={loading} sx={{ fontWeight: 700 }}>Cancel</Button>
                    <Button
                        variant="contained"
                        onClick={handleSave}
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
        </Box>
    );
};

export default PoliciesView;
