import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {DashboardComponent} from './pages/dashboard/dashboard.component';
import {DashboardComponent as AdminDashboardComponent} from './pages/admin/dashboard/dashboard.component';
import {UserAccessGuard} from './guards/user-access-guard.service';
import {LoginComponent} from './pages/auth/login/login.component';
import {LobbyEntryComponent} from './pages/lobby-entry/lobby-entry.component';
import {AdminAccessGuard} from './guards/admin-access-guard.service';
import {PasswordForgottenComponent} from './pages/auth/password-forgotten/password-forgotten.component';
import {SignupComponent} from './pages/auth/signup/signup.component';
import {VerifyComponent} from './pages/auth/verify/verify.component';

const routes: Routes = [
    {path: '', redirectTo: '/dashboard', pathMatch: 'full'},
    {path: 'dashboard', component: DashboardComponent, canActivate: [UserAccessGuard]},
    {path: 'admin/dashboard', component: AdminDashboardComponent, canActivate: [AdminAccessGuard]},
    {path: 'lobby/:spaceId/stream/:streamId', component: LobbyEntryComponent, canActivate: [UserAccessGuard]},
    {path: 'login', component: LoginComponent},
    {path: 'password-forgotten', component: PasswordForgottenComponent},
    {path: 'signup', component: SignupComponent},
    {path: 'verify', component: VerifyComponent}
];

@NgModule({
    imports: [
        RouterModule.forRoot(routes)
    ],
    exports: [RouterModule]
})
export class AppRoutingModule {
}
