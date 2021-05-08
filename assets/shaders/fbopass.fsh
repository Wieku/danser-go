#version 330
#extension GL_ARB_bindless_texture : enable

layout(bindless_sampler) uniform sampler2DArray tex;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture(tex, vec3(tex_coord, 0));
    color = in_color;
}