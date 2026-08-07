import React, {useState} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import ConfirmDialog from './confirm_dialog';

import {executeConfirm, hideConfirm, saveDesk, showConfirm} from '../actions';
import {labels} from '../labels';
import type {Desk} from '../types';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';

const AdminTab: React.FC = () => {
    const dispatch = useDispatch();
    const {desks, locations, adminSaving, confirmDialog} = useSelector((state: GlobalState) => getPluginStateFromGlobal(state));
    const defaultLocationId = locations[0]?.id || '';

    const [editingId, setEditingId] = useState<string | null>(null);
    const [form, setForm] = useState({
        name: '',
        sort_order: 0,
        is_active: true,
    });

    const resetForm = () => {
        setEditingId(null);
        setForm({name: '', sort_order: desks.length, is_active: true});
    };

    const startEdit = (desk: Desk) => {
        setEditingId(desk.id);
        setForm({
            name: desk.name,
            sort_order: desk.sort_order,
            is_active: desk.is_active,
        });
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!form.name.trim() || !defaultLocationId) {
            return;
        }
        saveDesk(dispatch, editingId, {
            location_id: defaultLocationId,
            name: form.name.trim(),
            sort_order: form.sort_order,
            is_active: form.is_active,
        }).then(() => resetForm());
    };

    const handleDelete = (desk: Desk) => {
        dispatch(showConfirm({
            action: 'delete_desk',
            deskId: desk.id,
            deskName: desk.name,
        }));
    };

    return (
        <div className='freedesk-admin-tab'>
            <h3 className='freedesk-admin-heading'>{editingId ? labels.deskEdit : labels.deskAdd}</h3>
            <form
                className='freedesk-admin-form'
                onSubmit={handleSubmit}
            >
                <label>
                    {labels.name}
                    <input
                        type='text'
                        value={form.name}
                        onChange={(e) => setForm({...form, name: e.target.value})}
                        maxLength={64}
                        required={true}
                    />
                </label>
                <label>
                    {labels.sortOrder}
                    <input
                        type='number'
                        value={form.sort_order}
                        onChange={(e) => setForm({...form, sort_order: Number(e.target.value)})}
                        min={0}
                    />
                </label>
                <label className='freedesk-admin-checkbox'>
                    <input
                        type='checkbox'
                        checked={form.is_active}
                        onChange={(e) => setForm({...form, is_active: e.target.checked})}
                    />
                    {labels.active}
                </label>
                <div className='freedesk-admin-form-actions'>
                    <button
                        type='submit'
                        className='btn btn-primary btn-sm'
                        disabled={adminSaving}
                    >
                        {editingId ? labels.update : labels.add}
                    </button>
                    {editingId && (
                        <button
                            type='button'
                            className='btn btn-tertiary btn-sm'
                            onClick={resetForm}
                        >
                            {labels.cancel}
                        </button>
                    )}
                </div>
            </form>

            <table className='freedesk-admin-table'>
                <thead>
                    <tr>
                        <th>{labels.name}</th>
                        <th>{labels.sortOrder}</th>
                        <th>{labels.status}</th>
                        <th>{labels.actions}</th>
                    </tr>
                </thead>
                <tbody>
                    {desks.map((desk) => (
                        <tr key={desk.id}>
                            <td>{desk.name}</td>
                            <td>{desk.sort_order}</td>
                            <td>{desk.is_active ? labels.active : labels.inactive}</td>
                            <td>
                                <button
                                    type='button'
                                    className='btn btn-tertiary btn-xs'
                                    onClick={() => startEdit(desk)}
                                >
                                    {labels.edit}
                                </button>
                                <button
                                    type='button'
                                    className='btn btn-tertiary btn-xs'
                                    onClick={() => handleDelete(desk)}
                                >
                                    {labels.delete}
                                </button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {confirmDialog && confirmDialog.action === 'delete_desk' && (
                <ConfirmDialog
                    dialog={confirmDialog}
                    onConfirm={() => executeConfirm(dispatch, confirmDialog, 0, 0)}
                    onCancel={() => dispatch(hideConfirm())}
                />
            )}
        </div>
    );
};

export default AdminTab;
