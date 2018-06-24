#version 330

in vec3 in_position;
in vec2 in_tex_coord;

uniform mat4 proj;
uniform mat4 model;

out vec2 tex_coord;
void main()
{
    gl_Position = proj * (model * vec4(in_position, 1));
    tex_coord = in_tex_coord;
}