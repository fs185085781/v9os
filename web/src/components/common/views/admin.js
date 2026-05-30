import { renderIcon } from "@/util/icon"
import { DataTable, UserRole,GatewayUserAccess,CloudSatelliteConfig,VirtualMachine,LicenseGlobal,LicenseThirdParty } from "@vicons/carbon";
import { Users } from "@vicons/tabler";
import {GroupsFilled,RememberMeOutlined,AppSettingsAltOutlined,PriceChangeOutlined,ListAltOutlined} from "@vicons/material";
import {ClusterOutlined} from "@vicons/antd";
import {PeopleList16Regular,PersonInfo16Regular} from "@vicons/fluent";
export function adminMenu() {
    const items = [
        {
            key: "base",
            icon: renderIcon(DataTable),
            children: [
                { key: "data", pkey: "system", icon: renderIcon(DataTable) },
                { key: "dead_msg", pkey: "system", icon: renderIcon(DataTable) },
                { key: "log", pkey: "system", icon: renderIcon(DataTable) },
                { key: "offline_chat_msg", pkey: "system", icon: renderIcon(DataTable) },
                { key: "plugin", pkey: "plugin", icon: renderIcon(DataTable) },
                { key: "plugin_column", pkey: "plugin", icon: renderIcon(DataTable) },
                { key: "plugin_data", pkey: "plugin", icon: renderIcon(DataTable) },
                { key: "plugin_table", pkey: "plugin", icon: renderIcon(DataTable) },
                { key: "plugin_web_data", pkey: "plugin", icon: renderIcon(DataTable) },
                { key: "user", pkey: "user", icon: renderIcon(DataTable) },
                { key: "user_settings", pkey: "user", icon: renderIcon(DataTable) },
                { key: "desktop_app", pkey: "user", icon: renderIcon(DataTable) },
            ],
        },
        {
            key: "user",
            icon: renderIcon(Users),
            children: [
                { key: "inface.user_ee.user", auth: "/api/user/page", icon: renderIcon(Users) },
                { key: "inface.user_ee.role", auth: "/api/role_ee/page", icon: renderIcon(UserRole) },
                { key: "inface.user_ee.dept", auth: "/api/dept_ee/page", icon: renderIcon(GroupsFilled) },
                { key: "inface.user_ee.auth", auth: "/api/auth_ee/page", icon: renderIcon(GatewayUserAccess) },
                { key: "inface.user_ee.config", auth: "/api/user_config_ee/list", icon: renderIcon(CloudSatelliteConfig) },
            ],
        },
        {
            key: "distributed",
            icon: renderIcon(ClusterOutlined),
            children: [
                { key: "inface.distributed_ee.machine", auth: "/api/machine_ee/list", icon: renderIcon(VirtualMachine) },
            ],
        },
        {
            key: "local_billing",
            icon: renderIcon(RememberMeOutlined),
            children: [
                { key: "inface.local_billing_ee.config", auth: "/api/local_billing_ee/config", icon: renderIcon(AppSettingsAltOutlined) },
                { key: "inface.local_billing_ee.price", auth: "/api/local_billing_ee/save_role_price", icon: renderIcon(PriceChangeOutlined) },
                { key: "inface.local_billing_ee.order", auth: "/api/local_billing_ee/orders", icon: renderIcon(ListAltOutlined) },
                { key: "inface.local_billing_ee.member", auth: "/api/local_billing_ee/members", icon: renderIcon(PeopleList16Regular) },
                { key: "inface.local_billing_ee.member_info", auth: "/api/local_billing_ee/member_page", icon: renderIcon(PersonInfo16Regular) },
            ],
        },
        {
            key: "official_license",
            icon: renderIcon(LicenseGlobal),
            children: [
                { key: "component.license.index",auth: "/api/license/page" , icon: renderIcon(LicenseThirdParty) },
            ],
        },
    ]
    return items
}